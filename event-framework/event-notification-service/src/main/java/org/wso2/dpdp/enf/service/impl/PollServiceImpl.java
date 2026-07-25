package org.wso2.dpdp.enf.service.impl;

import org.wso2.dpdp.enf.dao.DeliveryDAO;
import org.wso2.dpdp.enf.dao.EventDAO;
import org.wso2.dpdp.enf.dao.SubscriptionDAO;
import org.wso2.dpdp.enf.dao.TopicDAO;
import org.wso2.dpdp.enf.dao.impl.DeliveryDAOImpl;
import org.wso2.dpdp.enf.dao.impl.EventDAOImpl;
import org.wso2.dpdp.enf.dao.impl.SubscriptionDAOImpl;
import org.wso2.dpdp.enf.dao.impl.TopicDAOImpl;
import org.wso2.dpdp.enf.dao.model.Event;
import org.wso2.dpdp.enf.dao.model.PollDelivery;
import org.wso2.dpdp.enf.dao.model.Subscription;
import org.wso2.dpdp.enf.dao.model.Topic;
import org.wso2.dpdp.enf.service.PollService;
import org.wso2.dpdp.enf.service.dto.PollDeliveryItemDTO;
import org.wso2.dpdp.enf.service.dto.PollResponseDTO;
import org.wso2.dpdp.enf.service.exception.ENFException;
import org.wso2.dpdp.enf.service.security.HmacUtil;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

public class PollServiceImpl implements PollService {

    private static final int DEFAULT_MAX_EVENTS = 20;
    private static final int SYSTEM_MAX_EVENTS = 200;

    private final SubscriptionDAO subscriptionDAO;
    private final DeliveryDAO deliveryDAO;
    private final EventDAO eventDAO;
    private final TopicDAO topicDAO;

    public PollServiceImpl() {
        this.subscriptionDAO = new SubscriptionDAOImpl();
        this.deliveryDAO = new DeliveryDAOImpl();
        this.eventDAO = new EventDAOImpl();
        this.topicDAO = new TopicDAOImpl();
    }

    public PollServiceImpl(SubscriptionDAO subscriptionDAO, DeliveryDAO deliveryDAO, EventDAO eventDAO, TopicDAO topicDAO) {
        this.subscriptionDAO = subscriptionDAO;
        this.deliveryDAO = deliveryDAO;
        this.eventDAO = eventDAO;
        this.topicDAO = topicDAO;
    }

    @Override
    public PollResponseDTO pollEvents(String orgId, String groupId, String rawRequestBody, String signature,
                                      List<String> ackDeliveryIds, Map<String, Map<String, String>> setErrs,
                                      Integer maxEventsRequested, boolean returnImmediately) {

        // Find active pull/poll subscription for org & group using exact indexed query
        Optional<Subscription> pullSubOpt = subscriptionDAO.getActivePullSubscription(orgId, groupId);

        // Signature always required — reject if no active subscription found (prevents bypass)
        if (pullSubOpt.isEmpty()) {
            throw new ENFException("CS-4010", "Invalid signature",
                    "Unable to verify X-Event-Signature: no active pull/poll subscription found for this org and group.", 401);
        }
        String sharedSecret = pullSubOpt.get().getSharedSecret();
        if (!HmacUtil.verifySignature(rawRequestBody, sharedSecret, signature)) {
            throw new ENFException("CS-4010", "Invalid signature",
                    "X-Event-Signature is missing or does not match the HMAC computed from the subscription's shared secret.", 401);
        }

        // Process ACKs (Verify delivery ID belongs to this org for security)
        if (ackDeliveryIds != null && !ackDeliveryIds.isEmpty()) {
            for (String dlvId : ackDeliveryIds) {
                if (deliveryDAO.getPollDeliveryById(dlvId, orgId).isPresent()) {
                    deliveryDAO.updatePollDeliveryStatus(dlvId, "acknowledged");
                }
            }
        }

        // Process Errs (Verify delivery ID belongs to this org for security)
        if (setErrs != null && !setErrs.isEmpty()) {
            for (String dlvId : setErrs.keySet()) {
                if (deliveryDAO.getPollDeliveryById(dlvId, orgId).isPresent()) {
                    deliveryDAO.updatePollDeliveryStatus(dlvId, "err");
                }
            }
        }

        // Determine limit
        int limit = DEFAULT_MAX_EVENTS;
        if (maxEventsRequested != null && maxEventsRequested > 0) {
            limit = Math.min(maxEventsRequested, SYSTEM_MAX_EVENTS);
        }

        // Fetch pending poll items
        List<PollDelivery> pendingList = deliveryDAO.getPendingPollDeliveries(orgId, groupId, limit);
        List<PollDeliveryItemDTO> items = new ArrayList<>();

        for (PollDelivery pd : pendingList) {
            Optional<Event> eventOpt = eventDAO.getEventById(pd.getEventId());
            if (eventOpt.isPresent()) {
                Event event = eventOpt.get();
                Optional<Topic> topicOpt = topicDAO.getTopicById(event.getTopicId());
                String topicName = topicOpt.map(Topic::getName).orElse("unknown");

                Map<String, Object> payloadMap = parseJsonToMap(event.getPayload());

                items.add(new PollDeliveryItemDTO(
                        pd.getDeliveryId(),
                        event.getEventId(),
                        topicName,
                        event.getPurposes(),
                        payloadMap,
                        pd.getCreatedAt() != null ? pd.getCreatedAt().getTime() : System.currentTimeMillis()
                ));
            }
        }

        return new PollResponseDTO(items);
    }

    private Map<String, Object> parseJsonToMap(String json) {
        Map<String, Object> map = new HashMap<>();
        if (json == null || json.trim().isEmpty()) return map;
        map.put("raw", json);
        return map;
    }
}
