package org.wso2.dpdp.enf.service.impl;

import org.wso2.dpdp.enf.dao.DeliveryDAO;
import org.wso2.dpdp.enf.dao.EventDAO;
import org.wso2.dpdp.enf.dao.SubscriptionDAO;
import org.wso2.dpdp.enf.dao.impl.DeliveryDAOImpl;
import org.wso2.dpdp.enf.dao.impl.EventDAOImpl;
import org.wso2.dpdp.enf.dao.impl.SubscriptionDAOImpl;
import org.wso2.dpdp.enf.dao.model.Event;
import org.wso2.dpdp.enf.dao.model.Subscription;
import org.wso2.dpdp.enf.dao.model.WebhookDelivery;
import org.wso2.dpdp.enf.dao.model.WebhookDeliveryAudit;
import org.wso2.dpdp.enf.service.security.HmacUtil;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.sql.Timestamp;
import java.time.Duration;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.logging.Level;
import java.util.logging.Logger;

public class WebhookDeliveryProcessor {

    private static final Logger LOGGER = Logger.getLogger(WebhookDeliveryProcessor.class.getName());
    private static final int MAX_RETRIES = 5;
    private static final long BASE_BACKOFF_MILLIS = 10000; // 10 seconds base backoff

    private static final WebhookDeliveryProcessor INSTANCE = new WebhookDeliveryProcessor();

    private final DeliveryDAO deliveryDAO;
    private final SubscriptionDAO subscriptionDAO;
    private final EventDAO eventDAO;
    private final HttpClient httpClient;
    private ScheduledExecutorService scheduler;

    public static WebhookDeliveryProcessor getInstance() {
        return INSTANCE;
    }

    private WebhookDeliveryProcessor() {
        this(new DeliveryDAOImpl(), new SubscriptionDAOImpl(), new EventDAOImpl());
    }

    public WebhookDeliveryProcessor(DeliveryDAO deliveryDAO, SubscriptionDAO subscriptionDAO, EventDAO eventDAO) {
        this.deliveryDAO = deliveryDAO;
        this.subscriptionDAO = subscriptionDAO;
        this.eventDAO = eventDAO;
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(5))
                .build();
    }

    public synchronized void startScheduledJob() {
        if (scheduler != null && !scheduler.isShutdown()) {
            return;
        }
        scheduler = Executors.newSingleThreadScheduledExecutor();
        scheduler.scheduleWithFixedDelay(this::processPendingDeliveries, 5, 5, TimeUnit.SECONDS);
        LOGGER.info("Webhook Delivery Retry Scheduled Job started (running every 5 seconds).");
    }

    public synchronized void stopScheduledJob() {
        if (scheduler != null) {
            scheduler.shutdown();
            LOGGER.info("Webhook Delivery Retry Scheduled Job stopped.");
        }
    }

    public void processPendingDeliveries() {
        try {
            List<WebhookDelivery> pending = deliveryDAO.getPendingWebhookDeliveries(20);
            for (WebhookDelivery delivery : pending) {
                attemptDelivery(delivery);
            }
        } catch (Exception e) {
            LOGGER.log(Level.SEVERE, "Error in webhook delivery retry worker loop", e);
        }
    }

    public void attemptDelivery(WebhookDelivery delivery) {
        Optional<Event> eventOpt = eventDAO.getEventById(delivery.getEventId());
        if (eventOpt.isEmpty()) {
            LOGGER.warning("Event not found for delivery ID: " + delivery.getDeliveryId() + ". Marking as failed.");
            delivery.setStatus("failed");
            deliveryDAO.updateWebhookDeliveryStatus(delivery);
            return;
        }
        Event event = eventOpt.get();

        Optional<Subscription> subOpt = subscriptionDAO.getSubscriptionById(delivery.getSubscriptionId(), event.getOrgId());
        if (subOpt.isEmpty() || "deleted".equalsIgnoreCase(subOpt.get().getStatus())) {
            LOGGER.warning("Active subscription not found for delivery ID: " + delivery.getDeliveryId() + ". Marking as failed.");
            delivery.setStatus("failed");
            deliveryDAO.updateWebhookDeliveryStatus(delivery);
            return;
        }
        Subscription sub = subOpt.get();

        String callbackUrl = sub.getCallbackUrl();
        if (callbackUrl == null || callbackUrl.trim().isEmpty()) {
            delivery.setStatus("failed");
            deliveryDAO.updateWebhookDeliveryStatus(delivery);
            return;
        }

        String rawPayload = event.getPayload() != null ? event.getPayload() : "";
        String payload = formatPayloadWithDeliveryId(rawPayload, delivery.getDeliveryId());
        String signature = HmacUtil.calculateHmac(payload, sub.getSharedSecret());
        Timestamp now = new Timestamp(System.currentTimeMillis());

        String responseCode = null;
        boolean success = false;

        try {
            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(callbackUrl.trim()))
                    .header("Content-Type", "application/json")
                    .header("X-Event-Signature", signature)
                    .header("X-Org-Id", event.getOrgId())
                    .header("X-Event-Id", event.getEventId())
                    .header("X-Delivery-Id", delivery.getDeliveryId())
                    .timeout(Duration.ofSeconds(5))
                    .POST(HttpRequest.BodyPublishers.ofString(payload))
                    .build();

            HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
            int statusCode = response.statusCode();
            responseCode = String.valueOf(statusCode);

            if (statusCode >= 200 && statusCode < 300) {
                success = true;
            }
        } catch (Exception e) {
            responseCode = "ERR: " + (e.getMessage() != null ? e.getMessage() : e.getClass().getSimpleName());
            LOGGER.warning("HTTP delivery attempt failed for delivery ID " + delivery.getDeliveryId() + ": " + e.getMessage());
        }

        int newAttemptCount = delivery.getAttemptCount() + 1;
        delivery.setAttemptCount(newAttemptCount);

        if (success) {
            delivery.setStatus("delivered");
            delivery.setDeliveredAt(now);
            delivery.setNextRetryAt(null);
            LOGGER.info("Webhook delivery SUCCESS for ID: " + delivery.getDeliveryId() + " (Attempt #" + newAttemptCount + ")");
        } else {
            if (newAttemptCount >= MAX_RETRIES) {
                delivery.setStatus("failed");
                delivery.setNextRetryAt(null);
                LOGGER.warning("Webhook delivery FAILED after " + MAX_RETRIES + " attempts for ID: " + delivery.getDeliveryId());
            } else {
                delivery.setStatus("pending");
                // Exponential backoff: base * 2^(attempt-1)
                long delayMillis = (long) (BASE_BACKOFF_MILLIS * Math.pow(2, newAttemptCount - 1));
                Timestamp nextRetry = new Timestamp(now.getTime() + delayMillis);
                delivery.setNextRetryAt(nextRetry);
                LOGGER.info("Webhook delivery scheduled for RETRY (Attempt #" + (newAttemptCount + 1) + " in " + (delayMillis / 1000) + "s) for ID: " + delivery.getDeliveryId());
            }
        }

        deliveryDAO.updateWebhookDeliveryStatus(delivery);

        // Record Audit log
        WebhookDeliveryAudit audit = new WebhookDeliveryAudit(
                UUID.randomUUID().toString(),
                event.getEventId(),
                delivery.getDeliveryId(),
                event.getOrgId(),
                responseCode != null && responseCode.length() > 16 ? responseCode.substring(0, 16) : responseCode,
                now,
                now
        );
        deliveryDAO.addWebhookDeliveryAudit(audit);
    }

    /**
     * Safely injects deliveryId into the JSON payload using Jackson.
     * Prevents malformed JSON that could result from naive string concatenation.
     */
    private String formatPayloadWithDeliveryId(String rawPayload, String deliveryId) {
        try {
            ObjectMapper mapper = new ObjectMapper();
            java.util.Map<String, Object> payloadMap;
            if (rawPayload == null || rawPayload.trim().isEmpty()) {
                payloadMap = new java.util.LinkedHashMap<>();
            } else {
                payloadMap = mapper.readValue(rawPayload, new TypeReference<java.util.LinkedHashMap<String, Object>>() {});
            }
            // Insert deliveryId as the first field
            java.util.Map<String, Object> ordered = new java.util.LinkedHashMap<>();
            ordered.put("deliveryId", deliveryId);
            ordered.putAll(payloadMap);
            return mapper.writeValueAsString(ordered);
        } catch (Exception e) {
            LOGGER.warning("Failed to parse payload JSON for delivery " + deliveryId + "; using deliveryId-only payload. Cause: " + e.getMessage());
            return "{\"deliveryId\":\"" + deliveryId + "\"}";
        }
    }
}
