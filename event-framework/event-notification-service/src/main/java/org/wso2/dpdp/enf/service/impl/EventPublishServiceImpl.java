package org.wso2.dpdp.enf.service.impl;

import org.wso2.dpdp.enf.dao.DeliveryAckDAO;
import org.wso2.dpdp.enf.dao.DeliveryDAO;
import org.wso2.dpdp.enf.dao.EventDAO;
import org.wso2.dpdp.enf.dao.TopicDAO;
import org.wso2.dpdp.enf.dao.impl.DeliveryAckDAOImpl;
import org.wso2.dpdp.enf.dao.impl.DeliveryDAOImpl;
import org.wso2.dpdp.enf.dao.impl.EventDAOImpl;
import org.wso2.dpdp.enf.dao.impl.TopicDAOImpl;
import org.wso2.dpdp.enf.dao.model.Event;
import org.wso2.dpdp.enf.dao.model.PollDelivery;
import org.wso2.dpdp.enf.dao.model.SubscriptionDeliverySummary;
import org.wso2.dpdp.enf.dao.model.Topic;
import org.wso2.dpdp.enf.dao.model.WebhookDelivery;
import org.wso2.dpdp.enf.dao.model.WebhookDeliveryAck;
import org.wso2.dpdp.enf.dao.model.WebhookDeliveryAudit;
import org.wso2.dpdp.enf.service.EventFanOutService;
import org.wso2.dpdp.enf.service.EventPublishService;
import org.wso2.dpdp.enf.service.dto.EventDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionDeliveryAttemptDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionDeliveryDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionEventHistoryDTO;
import org.wso2.dpdp.enf.service.exception.ENFException;

import com.fasterxml.jackson.databind.ObjectMapper;

import java.sql.Timestamp;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import java.util.logging.Level;
import java.util.logging.Logger;

public class EventPublishServiceImpl implements EventPublishService {

    private static final Logger LOGGER = Logger.getLogger(EventPublishServiceImpl.class.getName());

    private final EventDAO eventDAO;
    private final TopicDAO topicDAO;
    private final EventFanOutService fanOutService;
    private final DeliveryDAO deliveryDAO;
    private final DeliveryAckDAO deliveryAckDAO;

    public EventPublishServiceImpl() {
        this(new EventDAOImpl(), new TopicDAOImpl(), new EventFanOutServiceImpl(), new DeliveryDAOImpl(), new DeliveryAckDAOImpl());
    }

    public EventPublishServiceImpl(EventDAO eventDAO, TopicDAO topicDAO, EventFanOutService fanOutService) {
        this(eventDAO, topicDAO, fanOutService, new DeliveryDAOImpl(), new DeliveryAckDAOImpl());
    }

    public EventPublishServiceImpl(EventDAO eventDAO, TopicDAO topicDAO, EventFanOutService fanOutService, DeliveryDAO deliveryDAO, DeliveryAckDAO deliveryAckDAO) {
        this.eventDAO = eventDAO;
        this.topicDAO = topicDAO;
        this.fanOutService = fanOutService;
        this.deliveryDAO = deliveryDAO;
        this.deliveryAckDAO = deliveryAckDAO;
    }

    @Override
    public EventDTO publishEvent(String orgId, String groupId, String topicName, List<String> purposes, Map<String, Object> payload) {
        if (topicName == null || topicName.trim().isEmpty() || payload == null) {
            throw new ENFException("CS-4001", "Malformed request", "Topic name and payload are required.", 400);
        }

        // Resolve Topic
        Optional<Topic> topicOpt = topicDAO.getTopicByOrgAndName(orgId, topicName.trim());
        if (topicOpt.isEmpty() || !"active".equalsIgnoreCase(topicOpt.get().getStatus())) {
            throw new ENFException("CS-4041", "Unknown topic", "The event referenced a topic name that does not exist for this org.", 404);
        }
        String topicId = topicOpt.get().getTopicId();

        String eventId = UUID.randomUUID().toString();
        Timestamp now = new Timestamp(System.currentTimeMillis());
        String payloadJson = convertPayloadToJson(payload);

        Event event = new Event(eventId, orgId, groupId, topicId, payloadJson, now);
        List<String> eventPurposes = purposes != null ? purposes : new ArrayList<>();
        event.setPurposes(eventPurposes);

        boolean added = eventDAO.addEvent(event);
        if (!added) {
            throw new ENFException("CS-5000", "Internal error", "Failed to persist event.", 500);
        }

        // Trigger Fan-Out synchronously; log errors but do not fail the event publication
        try {
            fanOutService.fanOutEvent(event, eventPurposes);
        } catch (Exception e) {
            LOGGER.log(Level.SEVERE, "Fan-out failed for event " + eventId + ". Deliveries may not have been created.", e);
        }

        long occurredAt = now.getTime();
        if (payload.containsKey("occurredAt") && payload.get("occurredAt") instanceof Number) {
            occurredAt = ((Number) payload.get("occurredAt")).longValue();
        }

        return new EventDTO(eventId, topicName.trim(), eventPurposes, payload, occurredAt, now.getTime());
    }

    private String convertPayloadToJson(Map<String, Object> map) {
        // Use Jackson's ObjectMapper for correct, safe JSON serialization
        // (handles escaping of \n, \t, \\, " and all other special characters)
        try {
            ObjectMapper mapper = new ObjectMapper();
            return mapper.writeValueAsString(map);
        } catch (Exception e) {
            LOGGER.log(Level.WARNING, "Failed to serialize payload to JSON using ObjectMapper; falling back to manual serialization.", e);
            // Fallback: safe manual build
            StringBuilder sb = new StringBuilder("{");
            boolean first = true;
            for (Map.Entry<String, Object> entry : map.entrySet()) {
                if (!first) sb.append(",");
                String escapedKey = entry.getKey().replace("\\", "\\\\").replace("\"", "\\\"");
                sb.append("\"").append(escapedKey).append("\":");
                Object val = entry.getValue();
                if (val == null) {
                    sb.append("null");
                } else if (val instanceof String s) {
                    // Escape all JSON special characters
                    String escaped = s.replace("\\", "\\\\").replace("\"", "\\\"").replace("\n", "\\n").replace("\r", "\\r").replace("\t", "\\t");
                    sb.append("\"").append(escaped).append("\"");
                } else {
                    sb.append(val);
                }
                first = false;
            }
            sb.append("}");
            return sb.toString();
        }
    }

    @Override
    public List<SubscriptionDeliveryDTO> listOrgEvents(String orgId, String statusFilter, String subscriptionIdFilter, String purposesFilter, String search, int limit, int offset, int[] totalOut) {
        if (orgId == null || orgId.trim().isEmpty()) {
            throw new ENFException("CS-4001", "Malformed request", "X-Org-Id header is required.", 400);
        }

        int lim = (limit <= 0) ? 20 : Math.min(limit, 100);
        int off = (offset < 0) ? 0 : offset;

        List<SubscriptionDeliverySummary> summaries = deliveryDAO.listOrgDeliveries(orgId.trim(), statusFilter, subscriptionIdFilter, purposesFilter, search, lim, off, totalOut);
        List<SubscriptionDeliveryDTO> dtoList = new ArrayList<>();
        for (SubscriptionDeliverySummary summary : summaries) {
            dtoList.add(new SubscriptionDeliveryDTO(
                    summary.getDeliveryId(),
                    summary.getEventId(),
                    summary.getTopicName(),
                    summary.getCurrentStatus() != null ? summary.getCurrentStatus().toUpperCase() : "PENDING",
                    summary.getDeliveryMode() != null ? summary.getDeliveryMode().toUpperCase() : "WEBHOOK",
                    summary.getOccurredAt() != null ? summary.getOccurredAt().getTime() : (summary.getCreatedAt() != null ? summary.getCreatedAt().getTime() : System.currentTimeMillis()),
                    summary.getPayload()
            ));
        }
        return dtoList;
    }

    @Override
    public SubscriptionEventHistoryDTO getOrgEventHistory(String orgId, String deliveryId) {
        if (orgId == null || orgId.trim().isEmpty() || deliveryId == null || deliveryId.trim().isEmpty()) {
            throw new ENFException("CS-4001", "Malformed request", "X-Org-Id and deliveryId are required.", 400);
        }

        Optional<SubscriptionDeliverySummary> summaryOpt = deliveryDAO.getOrgDeliveryById(orgId.trim(), deliveryId.trim());
        if (summaryOpt.isEmpty()) {
            throw new ENFException("CS-4042", "Delivery not found", "Event delivery with ID '" + deliveryId + "' does not exist for this org.", 404);
        }

        SubscriptionDeliverySummary summary = summaryOpt.get();
        String mode = summary.getDeliveryMode() != null ? summary.getDeliveryMode().toUpperCase() : "WEBHOOK";

        SubscriptionEventHistoryDTO dto = new SubscriptionEventHistoryDTO(
                summary.getDeliveryId(),
                summary.getEventId(),
                summary.getTopicName(),
                mode,
                summary.getCurrentStatus() != null ? summary.getCurrentStatus().toUpperCase() : "PENDING",
                summary.getOccurredAt() != null ? summary.getOccurredAt().getTime() : (summary.getCreatedAt() != null ? summary.getCreatedAt().getTime() : System.currentTimeMillis()),
                summary.getPayload()
        );

        if ("WEBHOOK".equals(mode)) {
            Optional<WebhookDelivery> whOpt = deliveryDAO.getWebhookDeliveryById(deliveryId.trim(), orgId.trim());
            if (whOpt.isPresent()) {
                WebhookDelivery wh = whOpt.get();
                if (wh.getNextRetryAt() != null) {
                    dto.setNextRetryAt(wh.getNextRetryAt().getTime());
                }
            }

            Optional<WebhookDeliveryAck> ackOpt = deliveryAckDAO.getDeliveryAckByDeliveryId(deliveryId.trim());
            if (ackOpt.isPresent()) {
                WebhookDeliveryAck ack = ackOpt.get();
                dto.setCompletionStatus(ack.getCompletionStatus() != null ? ack.getCompletionStatus().toUpperCase() : "COMPLETED");
                dto.setCompletionEvidence(ack.getCompletionEvidence());
            }

            List<WebhookDeliveryAudit> audits = deliveryDAO.getWebhookDeliveryAudits(deliveryId.trim());
            List<SubscriptionDeliveryAttemptDTO> attempts = new ArrayList<>();
            int attemptNumber = 1;
            for (WebhookDeliveryAudit audit : audits) {
                Integer httpStatus = null;
                String error = null;
                String status = "FAILED";
                if (audit.getResponseCode() != null) {
                    try {
                        httpStatus = Integer.parseInt(audit.getResponseCode().trim());
                        if (httpStatus >= 200 && httpStatus < 300) {
                            status = "DELIVERED";
                        } else {
                            error = "HTTP " + httpStatus;
                        }
                    } catch (NumberFormatException e) {
                        error = audit.getResponseCode();
                    }
                }
                attempts.add(new SubscriptionDeliveryAttemptDTO(
                        attemptNumber++,
                        status,
                        audit.getAttemptAt() != null ? audit.getAttemptAt().getTime() : (audit.getCreatedAt() != null ? audit.getCreatedAt().getTime() : System.currentTimeMillis()),
                        httpStatus,
                        error
                ));
            }
            if (attempts.isEmpty()) {
                attempts.add(new SubscriptionDeliveryAttemptDTO(
                        1,
                        summary.getCurrentStatus() != null ? summary.getCurrentStatus().toUpperCase() : "PENDING",
                        summary.getCreatedAt() != null ? summary.getCreatedAt().getTime() : System.currentTimeMillis(),
                        null,
                        null
                ));
            }
            dto.setHistory(attempts);
        } else {
            Optional<PollDelivery> pollOpt = deliveryDAO.getPollDeliveryById(deliveryId.trim(), orgId.trim());
            List<SubscriptionDeliveryAttemptDTO> attempts = new ArrayList<>();
            String pollStatus = summary.getCurrentStatus() != null ? summary.getCurrentStatus().toUpperCase() : "PENDING";
            long timestamp = summary.getOccurredAt() != null ? summary.getOccurredAt().getTime() : System.currentTimeMillis();

            if (pollOpt.isPresent()) {
                PollDelivery pd = pollOpt.get();
                if (pd.getCompletedAt() != null) {
                    timestamp = pd.getCompletedAt().getTime();
                }
            }

            dto.setCompletionStatus(pollStatus);
            attempts.add(new SubscriptionDeliveryAttemptDTO(
                    1,
                    pollStatus,
                    timestamp,
                    null,
                    null
            ));
            dto.setHistory(attempts);
        }

        return dto;
    }
}
