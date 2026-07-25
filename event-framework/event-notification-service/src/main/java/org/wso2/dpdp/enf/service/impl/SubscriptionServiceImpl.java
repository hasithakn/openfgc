package org.wso2.dpdp.enf.service.impl;

import org.wso2.dpdp.enf.dao.DeliveryAckDAO;
import org.wso2.dpdp.enf.dao.DeliveryDAO;
import org.wso2.dpdp.enf.dao.SubscriptionDAO;
import org.wso2.dpdp.enf.dao.TopicDAO;
import org.wso2.dpdp.enf.dao.impl.DeliveryAckDAOImpl;
import org.wso2.dpdp.enf.dao.impl.DeliveryDAOImpl;
import org.wso2.dpdp.enf.dao.impl.SubscriptionDAOImpl;
import org.wso2.dpdp.enf.dao.impl.TopicDAOImpl;
import org.wso2.dpdp.enf.dao.model.PollDelivery;
import org.wso2.dpdp.enf.dao.model.Subscription;
import org.wso2.dpdp.enf.dao.model.SubscriptionDeliverySummary;
import org.wso2.dpdp.enf.dao.model.Topic;
import org.wso2.dpdp.enf.dao.model.WebhookDelivery;
import org.wso2.dpdp.enf.dao.model.WebhookDeliveryAck;
import org.wso2.dpdp.enf.dao.model.WebhookDeliveryAudit;
import org.wso2.dpdp.enf.service.SubscriptionService;
import org.wso2.dpdp.enf.service.dto.DeliveryConfigDTO;
import org.wso2.dpdp.enf.service.dto.FilterDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionDeliveryAttemptDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionDeliveryDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionEventHistoryDTO;
import org.wso2.dpdp.enf.service.exception.ENFException;

import java.net.URI;
import java.net.URLEncoder;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.sql.Timestamp;
import java.time.Duration;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashSet;
import java.util.List;
import java.util.Optional;
import java.util.Set;
import java.util.UUID;

public class SubscriptionServiceImpl implements SubscriptionService {

    private final SubscriptionDAO subscriptionDAO;
    private final TopicDAO topicDAO;
    private final DeliveryDAO deliveryDAO;
    private final DeliveryAckDAO deliveryAckDAO;

    public SubscriptionServiceImpl() {
        this(new SubscriptionDAOImpl(), new TopicDAOImpl(), new DeliveryDAOImpl(), new DeliveryAckDAOImpl());
    }

    public SubscriptionServiceImpl(SubscriptionDAO subscriptionDAO, TopicDAO topicDAO) {
        this(subscriptionDAO, topicDAO, new DeliveryDAOImpl(), new DeliveryAckDAOImpl());
    }

    public SubscriptionServiceImpl(SubscriptionDAO subscriptionDAO, TopicDAO topicDAO, DeliveryDAO deliveryDAO, DeliveryAckDAO deliveryAckDAO) {
        this.subscriptionDAO = subscriptionDAO;
        this.topicDAO = topicDAO;
        this.deliveryDAO = deliveryDAO;
        this.deliveryAckDAO = deliveryAckDAO;
    }

    @Override
    public SubscriptionDTO createSubscription(String orgId, String groupId, String topicName, FilterDTO filter, DeliveryConfigDTO delivery) {
        if (topicName == null || topicName.trim().isEmpty() || filter == null || filter.getType() == null || delivery == null || delivery.getMode() == null) {
            throw new ENFException("CS-4001", "Malformed request", "Required subscription fields are missing.", 400);
        }

        String filterType = filter.getType().trim().toLowerCase().replaceAll("-", "_");
        if (!"all".equals(filterType) && !"specific".equals(filterType) && !"all_except".equals(filterType)) {
            throw new ENFException("CS-4002", "Validation failed", "filter.type must be one of 'all', 'specific', or 'all_except'. Received: '" + filter.getType() + "'", 422);
        }

        List<String> purposes = filter.getPurposes() != null ? new ArrayList<>(filter.getPurposes()) : new ArrayList<>();

        // Validation rule 422 CS-4002: filter.purposes required for all_except / specific
        if (("all_except".equals(filterType) || "specific".equals(filterType)) && purposes.isEmpty()) {
            throw new ENFException("CS-4002", "Validation failed", "filter.purposes is required for filter.type all_except/specific.", 422);
        }

        String deliveryMode = delivery.getMode().trim().toLowerCase();
        if (!"webhook".equals(deliveryMode) && !"poll".equals(deliveryMode) && !"pull".equals(deliveryMode)) {
            throw new ENFException("CS-4002", "Validation failed", "delivery.mode must be one of 'webhook', 'poll', or 'pull'. Received: '" + delivery.getMode() + "'", 422);
        }

        // Validation rule 422 CS-4002: callbackUrl required for webhook
        if ("webhook".equals(deliveryMode) && (delivery.getCallbackUrl() == null || delivery.getCallbackUrl().trim().isEmpty())) {
            throw new ENFException("CS-4002", "Validation failed", "delivery.callbackUrl is required for delivery.mode webhook.", 422);
        }

        if (delivery.getSharedSecret() == null || delivery.getSharedSecret().trim().isEmpty()) {
            throw new ENFException("CS-4001", "Malformed request", "delivery.sharedSecret is required.", 400);
        }

        // Resolve Topic
        Optional<Topic> topicOpt = topicDAO.getTopicByOrgAndName(orgId, topicName.trim());
        if (topicOpt.isEmpty() || !"active".equalsIgnoreCase(topicOpt.get().getStatus())) {
            throw new ENFException("CS-4040", "Topic not found", "No active topic exists with this name for the given org.", 404);
        }
        String topicId = topicOpt.get().getTopicId();

        // Sort purposes to guarantee uniqueness matching
        Collections.sort(purposes);

        // Check for duplicate active subscription
        Optional<Subscription> duplicate = subscriptionDAO.findDuplicateSubscription(orgId, groupId, topicId, filterType, purposes);
        if (duplicate.isPresent()) {
            Subscription existing = duplicate.get();
            SubscriptionDTO existingDTO = new SubscriptionDTO();
            existingDTO.setSubscriptionId(existing.getSubscriptionId());
            existingDTO.setAlreadyExists(true);
            existingDTO.setMessage("Subscription already exists for topic '" + topicName + "' and the specified purposes.");
            return existingDTO;
        }

        // Check for conflicting/overlapping active subscriptions
        validateSubscriptionConflict(orgId, groupId, topicId, filterType, purposes);

        String subId = UUID.randomUUID().toString();
        Timestamp now = new Timestamp(System.currentTimeMillis());
        String initialStatus = "webhook".equals(deliveryMode) ? "pending" : "active";

        Subscription sub = new Subscription(
                subId,
                orgId,
                groupId,
                topicId,
                initialStatus,
                filterType,
                "webhook".equals(deliveryMode) ? delivery.getCallbackUrl().trim() : null,
                delivery.getSharedSecret().trim(),
                now,
                now,
                deliveryMode
        );
        sub.setPurposes(purposes);

        boolean created = subscriptionDAO.addSubscription(sub);
        if (!created) {
            throw new ENFException("CS-5000", "Internal error", "Failed to persist subscription.", 500);
        }

        // Perform Webhook Intent Verification if delivery mode is webhook
        if ("webhook".equals(deliveryMode)) {
            try {
                verifyWebhookCallback(delivery.getCallbackUrl().trim(), topicName);
                subscriptionDAO.updateSubscriptionStatus(subId, "active");
                sub.setStatus("active");
            } catch (Exception e) {
                subscriptionDAO.updateSubscriptionStatus(subId, "stale");
                sub.setStatus("stale");
                String detailMsg = e instanceof ENFException ? e.getMessage() : e.getMessage();
                throw new ENFException("CS-4220", "Webhook verification failed",
                        "Failed to verify webhook callback URL [" + delivery.getCallbackUrl().trim() + "]. Subscription status set to 'stale'. Detail: " + detailMsg, 422);
            }
        }

        return mapToDTO(sub, topicName);
    }

    @Override
    public List<SubscriptionDTO> listSubscriptions(String orgId, String status, String purposes, String search, int limit, int offset, String sort, int[] totalOut) {
        List<Subscription> subs = subscriptionDAO.listSubscriptions(orgId, status, purposes, search, limit, offset, sort, totalOut);
        List<SubscriptionDTO> result = new ArrayList<>();
        for (Subscription sub : subs) {
            Optional<Topic> topicOpt = topicDAO.getTopicById(sub.getTopicId());
            String topicName = topicOpt.map(Topic::getName).orElse("unknown");
            result.add(mapToDTO(sub, topicName));
        }
        return result;
    }

    @Override
    public SubscriptionDTO getSubscription(String orgId, String subscriptionId) {
        Optional<Subscription> subOpt = subscriptionDAO.getSubscriptionById(subscriptionId, orgId);
        if (subOpt.isEmpty() || "deleted".equalsIgnoreCase(subOpt.get().getStatus())) {
            throw new ENFException("CS-4040", "Resource not found", "No subscription exists with this ID for the given org.", 404);
        }
        Subscription sub = subOpt.get();
        Optional<Topic> topicOpt = topicDAO.getTopicById(sub.getTopicId());
        String topicName = topicOpt.map(Topic::getName).orElse("unknown");
        return mapToDTO(sub, topicName);
    }

    @Override
    public boolean deleteSubscription(String orgId, String subscriptionId) {
        Optional<Subscription> subOpt = subscriptionDAO.getSubscriptionById(subscriptionId, orgId);
        if (subOpt.isEmpty() || "deleted".equalsIgnoreCase(subOpt.get().getStatus())) {
            throw new ENFException("CS-4040", "Resource not found", "No subscription exists with this ID for the given org.", 404);
        }

        // Check if undelivered WEBHOOK_DELIVERY or POLL_DELIVERY rows exist
        if (subscriptionDAO.hasPendingDeliveries(subscriptionId)) {
            throw new ENFException("CS-4092", "Subscription has active deliveries", "This subscription cannot be deleted while undelivered WEBHOOK_DELIVERY or POLL_DELIVERY rows exist for it.", 409);
        }

        return subscriptionDAO.updateSubscriptionStatus(subscriptionId, "deleted");
    }

    private void validateSubscriptionConflict(String orgId, String groupId, String topicId, String newFilterType, List<String> newPurposes) {
        List<Subscription> activeSubs = subscriptionDAO.getActiveSubscriptionsForMatching(orgId, groupId, topicId);
        if (activeSubs == null || activeSubs.isEmpty()) {
            return;
        }

        Set<String> newPurposeSet = new HashSet<>(newPurposes != null ? newPurposes : Collections.emptyList());

        for (Subscription sub : activeSubs) {
            String existMode = sub.getPurposeFilterMode().toLowerCase();
            Set<String> existPurposes = new HashSet<>(sub.getPurposes() != null ? sub.getPurposes() : Collections.emptyList());

            // Case A: New or existing filter mode is 'all'
            if ("all".equals(newFilterType) || "all".equals(existMode)) {
                throw new ENFException("CS-4090", "Subscription conflict",
                        "A subscription with filter type 'all' conflicts with other subscriptions for this topic.", 409);
            }

            // Case B: New filter mode is 'specific'
            if ("specific".equals(newFilterType)) {
                if ("specific".equals(existMode)) {
                    Set<String> overlap = new HashSet<>(newPurposeSet);
                    overlap.retainAll(existPurposes);
                    if (!overlap.isEmpty()) {
                        throw new ENFException("CS-4090", "Subscription conflict",
                                "Specific subscription purpose(s) " + overlap + " overlap with an existing specific subscription.", 409);
                    }
                } else if ("all_except".equals(existMode)) {
                    Set<String> nonExcluded = new HashSet<>(newPurposeSet);
                    nonExcluded.removeAll(existPurposes);
                    if (!nonExcluded.isEmpty()) {
                        throw new ENFException("CS-4090", "Subscription conflict",
                                "Specific subscription purpose(s) " + nonExcluded + " are not excluded by the existing all_except subscription.", 409);
                    }
                }
            }

            // Case C: New filter mode is 'all_except'
            if ("all_except".equals(newFilterType)) {
                if ("all_except".equals(existMode)) {
                    throw new ENFException("CS-4090", "Subscription conflict",
                            "An all_except subscription already exists for this subscriber and topic.", 409);
                } else if ("specific".equals(existMode)) {
                    Set<String> nonExcluded = new HashSet<>(existPurposes);
                    nonExcluded.removeAll(newPurposeSet);
                    if (!nonExcluded.isEmpty()) {
                        throw new ENFException("CS-4090", "Subscription conflict",
                                "New all_except subscription does not exclude existing specific purpose(s) " + nonExcluded + ".", 409);
                    }
                }
            }
        }
    }

    private void verifyWebhookCallback(String callbackUrl, String topicName) {
        // --- SSRF Prevention: validate scheme and disallow private/loopback addresses ---
        try {
            URI parsedUri = URI.create(callbackUrl);
            String scheme = parsedUri.getScheme();
            if (scheme == null || (!scheme.equalsIgnoreCase("http") && !scheme.equalsIgnoreCase("https"))) {
                throw new ENFException("CS-4220", "Webhook verification failed",
                        "Callback URL must use http or https scheme. Received: '" + scheme + "'", 422);
            }
            String host = parsedUri.getHost();
            if (host == null || host.trim().isEmpty()) {
                throw new ENFException("CS-4220", "Webhook verification failed",
                        "Callback URL does not contain a valid host.", 422);
            }
            // Block loopback and private/internal network ranges (SSRF protection)
            // Allow localhost explicitly only during development — remove in production
            java.net.InetAddress addr;
            try {
                addr = java.net.InetAddress.getByName(host);
            } catch (java.net.UnknownHostException e) {
                throw new ENFException("CS-4220", "Webhook verification failed",
                        "Callback URL host cannot be resolved: '" + host + "'", 422);
            }
            if (addr.isLoopbackAddress() || addr.isSiteLocalAddress() || addr.isLinkLocalAddress() || addr.isAnyLocalAddress()) {
                // For local development purposes, loopback is allowed. Comment out the next line in production:
                // throw new ENFException("CS-4220", "Webhook verification failed",
                //     "Callback URL must not resolve to a private or loopback address.", 422);
            }
        } catch (ENFException e) {
            throw e;
        } catch (IllegalArgumentException e) {
            throw new ENFException("CS-4220", "Webhook verification failed",
                    "Callback URL is not a valid URI: '" + callbackUrl + "'", 422);
        }

        String challenge = UUID.randomUUID().toString();
        String delimiter = callbackUrl.contains("?") ? "&" : "?";
        String verificationUrl = callbackUrl + delimiter
                + "hub.mode=subscribe"
                + "&hub.topic=" + URLEncoder.encode(topicName, StandardCharsets.UTF_8)
                + "&hub.challenge=" + challenge
                + "&challenge=" + challenge;

        try {
            HttpClient client = HttpClient.newBuilder()
                    .connectTimeout(Duration.ofSeconds(5))
                    .build();

            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(verificationUrl))
                    .timeout(Duration.ofSeconds(5))
                    .GET()
                    .build();

            HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());

            if (response.statusCode() != 200) {
                throw new ENFException("CS-4220", "Webhook verification failed",
                        "Callback URL [" + callbackUrl + "] responded with HTTP " + response.statusCode() + " during intent verification.", 422);
            }

            String body = response.body() != null ? response.body().trim() : "";
            if (!body.contains(challenge)) {
                throw new ENFException("CS-4220", "Webhook verification failed",
                        "Callback URL [" + callbackUrl + "] did not echo back the expected challenge string. Received: '" + body + "'", 422);
            }
        } catch (ENFException e) {
            throw e;
        } catch (Exception e) {
            throw new ENFException("CS-4220", "Webhook verification failed",
                    "Failed to connect to webhook callback URL [" + callbackUrl + "]: " + e.getMessage(), 422);
        }
    }

    private SubscriptionDTO mapToDTO(Subscription sub, String topicName) {
        FilterDTO filterDTO = new FilterDTO(sub.getPurposeFilterMode(), sub.getPurposes());
        DeliveryConfigDTO deliveryConfigDTO = new DeliveryConfigDTO(sub.getDeliveryMode(), sub.getCallbackUrl(), null); // sharedSecret writeOnly
        return new SubscriptionDTO(
                sub.getSubscriptionId(),
                topicName,
                filterDTO,
                deliveryConfigDTO,
                sub.getStatus(),
                sub.getCreatedAt() != null ? sub.getCreatedAt().getTime() : System.currentTimeMillis(),
                sub.getUpdatedAt() != null ? sub.getUpdatedAt().getTime() : System.currentTimeMillis()
        );
    }

    @Override
    public List<SubscriptionDeliveryDTO> listSubscriptionEvents(String orgId, String subscriptionId, int limit, int offset, int[] totalOut) {
        if (subscriptionId == null || subscriptionId.trim().isEmpty()) {
            throw new ENFException("CS-4001", "Malformed request", "subscriptionId is required.", 400);
        }

        // Validate subscription exists and belongs to orgId
        Optional<Subscription> subOpt = subscriptionDAO.getSubscriptionById(subscriptionId.trim(), orgId);
        if (subOpt.isEmpty()) {
            throw new ENFException("CS-4040", "Subscription not found", "Subscription with ID '" + subscriptionId + "' does not exist for this org.", 404);
        }

        int lim = (limit <= 0) ? 20 : Math.min(limit, 100);
        int off = (offset < 0) ? 0 : offset;

        List<SubscriptionDeliverySummary> summaries = deliveryDAO.listSubscriptionDeliveries(subscriptionId.trim(), lim, off, totalOut);
        List<SubscriptionDeliveryDTO> dtoList = new ArrayList<>();
        for (SubscriptionDeliverySummary summary : summaries) {
            dtoList.add(new SubscriptionDeliveryDTO(
                    summary.getDeliveryId(),
                    summary.getEventId(),
                    summary.getTopicName(),
                    summary.getCurrentStatus() != null ? summary.getCurrentStatus().toUpperCase() : "PENDING",
                    summary.getDeliveryMode() != null ? summary.getDeliveryMode().toUpperCase() : "WEBHOOK",
                    summary.getOccurredAt() != null ? summary.getOccurredAt().getTime() : (summary.getCreatedAt() != null ? summary.getCreatedAt().getTime() : System.currentTimeMillis())
            ));
        }
        return dtoList;
    }

    @Override
    public SubscriptionEventHistoryDTO getSubscriptionEventHistory(String orgId, String subscriptionId, String deliveryId) {
        if (subscriptionId == null || subscriptionId.trim().isEmpty() || deliveryId == null || deliveryId.trim().isEmpty()) {
            throw new ENFException("CS-4001", "Malformed request", "subscriptionId and deliveryId are required.", 400);
        }

        // Validate subscription exists and belongs to orgId
        Optional<Subscription> subOpt = subscriptionDAO.getSubscriptionById(subscriptionId.trim(), orgId);
        if (subOpt.isEmpty()) {
            throw new ENFException("CS-4040", "Subscription not found", "Subscription with ID '" + subscriptionId + "' does not exist for this org.", 404);
        }

        // Validate delivery exists and belongs to this subscription
        Optional<SubscriptionDeliverySummary> summaryOpt = deliveryDAO.getSubscriptionDeliveryById(subscriptionId.trim(), deliveryId.trim());
        if (summaryOpt.isEmpty()) {
            throw new ENFException("CS-4042", "Delivery not found", "Delivery with ID '" + deliveryId + "' does not exist for subscription '" + subscriptionId + "'.", 404);
        }

        SubscriptionDeliverySummary summary = summaryOpt.get();
        String mode = summary.getDeliveryMode() != null ? summary.getDeliveryMode().toUpperCase() : "WEBHOOK";

        SubscriptionEventHistoryDTO dto = new SubscriptionEventHistoryDTO(
                summary.getDeliveryId(),
                summary.getEventId(),
                summary.getTopicName(),
                mode,
                summary.getCurrentStatus() != null ? summary.getCurrentStatus().toUpperCase() : "PENDING",
                summary.getOccurredAt() != null ? summary.getOccurredAt().getTime() : (summary.getCreatedAt() != null ? summary.getCreatedAt().getTime() : System.currentTimeMillis())
        );

        if ("WEBHOOK".equals(mode)) {
            // Webhook details
            Optional<WebhookDelivery> whOpt = deliveryDAO.getWebhookDeliveryById(deliveryId.trim(), orgId);
            if (whOpt.isPresent()) {
                WebhookDelivery wh = whOpt.get();
                if (wh.getNextRetryAt() != null) {
                    dto.setNextRetryAt(wh.getNextRetryAt().getTime());
                }
            }

            // Webhook ACK if present
            Optional<WebhookDeliveryAck> ackOpt = deliveryAckDAO.getDeliveryAckByDeliveryId(deliveryId.trim());
            if (ackOpt.isPresent()) {
                WebhookDeliveryAck ack = ackOpt.get();
                dto.setCompletionStatus(ack.getCompletionStatus() != null ? ack.getCompletionStatus().toUpperCase() : "COMPLETED");
                dto.setCompletionEvidence(ack.getCompletionEvidence());
            }

            // Webhook attempt audit history
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
            // If no audits found yet but webhook delivery exists (initial attempt pending)
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
            // Poll details
            Optional<PollDelivery> pollOpt = deliveryDAO.getPollDeliveryById(deliveryId.trim(), orgId);
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
