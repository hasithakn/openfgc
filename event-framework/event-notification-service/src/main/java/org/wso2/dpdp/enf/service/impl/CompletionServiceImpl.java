package org.wso2.dpdp.enf.service.impl;

import org.wso2.dpdp.enf.dao.DeliveryAckDAO;
import org.wso2.dpdp.enf.dao.DeliveryDAO;
import org.wso2.dpdp.enf.dao.SubscriptionDAO;
import org.wso2.dpdp.enf.dao.impl.DeliveryAckDAOImpl;
import org.wso2.dpdp.enf.dao.impl.DeliveryDAOImpl;
import org.wso2.dpdp.enf.dao.impl.SubscriptionDAOImpl;
import org.wso2.dpdp.enf.dao.model.Subscription;
import org.wso2.dpdp.enf.dao.model.WebhookDelivery;
import org.wso2.dpdp.enf.dao.model.WebhookDeliveryAck;
import org.wso2.dpdp.enf.service.CompletionService;
import org.wso2.dpdp.enf.service.dto.CompletionAckDTO;
import org.wso2.dpdp.enf.service.exception.ENFException;
import org.wso2.dpdp.enf.service.security.HmacUtil;

import java.sql.Timestamp;
import java.util.Optional;
import java.util.UUID;

public class CompletionServiceImpl implements CompletionService {

    private final DeliveryDAO deliveryDAO;
    private final DeliveryAckDAO deliveryAckDAO;
    private final SubscriptionDAO subscriptionDAO;

    public CompletionServiceImpl() {
        this.deliveryDAO = new DeliveryDAOImpl();
        this.deliveryAckDAO = new DeliveryAckDAOImpl();
        this.subscriptionDAO = new SubscriptionDAOImpl();
    }

    public CompletionServiceImpl(DeliveryDAO deliveryDAO, DeliveryAckDAO deliveryAckDAO,
            SubscriptionDAO subscriptionDAO) {
        this.deliveryDAO = deliveryDAO;
        this.deliveryAckDAO = deliveryAckDAO;
        this.subscriptionDAO = subscriptionDAO;
    }

    @Override
    public CompletionAckDTO submitCompletion(String orgId, String groupId, String deliveryId,
            String rawRequestBody, String signature,
            String completionStatus, String completionEvidence, Long completedAt) {

        if (completionStatus == null || completionEvidence == null || completionEvidence.trim().isEmpty()) {
            throw new ENFException("CS-4001", "Malformed request",
                    "completionStatus and completionEvidence are required.", 400);
        }

        String normStatus = completionStatus.trim().toLowerCase();
        if (!"completed".equals(normStatus) && !"disputed".equals(normStatus) && !"partial".equals(normStatus)) {
            throw new ENFException("CS-4002", "Validation failed",
                    "completionStatus must be one of 'completed', 'disputed', or 'partial'. Received: '" + completionStatus + "'", 422);
        }
        completionStatus = normStatus;

        // Check delivery existence in WEBHOOK_DELIVERY
        Optional<WebhookDelivery> webhookDeliveryOpt = deliveryDAO.getWebhookDeliveryById(deliveryId, orgId);
        if (webhookDeliveryOpt.isEmpty()) {
            throw new ENFException("CS-4042", "Delivery not found",
                    "No WEBHOOK_DELIVERY exists with this deliveryId for the given org.", 404);
        }

        WebhookDelivery webhookDelivery = webhookDeliveryOpt.get();
        String subscriptionId = webhookDelivery.getSubscriptionId();

        // Retrieve Subscription for shared secret verification — always required
        Optional<Subscription> subOpt = subscriptionDAO.getSubscriptionById(subscriptionId, orgId);
        if (subOpt.isEmpty()) {
            // Subscription was deleted or not found; still reject — do not allow bypass
            throw new ENFException("CS-4010", "Invalid signature",
                    "Unable to verify X-Event-Signature: no active subscription found for this delivery.", 401);
        }
        String sharedSecret = subOpt.get().getSharedSecret();
        if (!HmacUtil.verifySignature(rawRequestBody, sharedSecret, signature)) {
            throw new ENFException("CS-4010", "Invalid signature",
                    "X-Event-Signature is missing or does not match the HMAC computed from the subscription's shared secret.",
                    401);
        }

        // Check if completion report already exists (UQ_EDA_DELIVERY)
        Optional<WebhookDeliveryAck> existingAck = deliveryAckDAO.getDeliveryAckByDeliveryId(deliveryId);
        if (existingAck.isPresent()) {
            throw new ENFException("CS-4093", "Completion already recorded",
                    "A WEBHOOK_DELIVERY_ACK already exists for this deliveryId; submit a dispute through support instead of resubmitting.",
                    409);
        }

        String ackId = UUID.randomUUID().toString();
        Timestamp completedTimestamp = completedAt != null ? new Timestamp(completedAt)
                : new Timestamp(System.currentTimeMillis());

        WebhookDeliveryAck ack = new WebhookDeliveryAck(
                ackId,
                deliveryId,
                completedTimestamp,
                completionStatus,
                completionEvidence.trim());

        boolean saved = deliveryAckDAO.addDeliveryAck(ack);
        if (!saved) {
            throw new ENFException("CS-5000", "Internal error", "Failed to record completion report.", 500);
        }

        webhookDelivery.setStatus("completed");
        deliveryDAO.updateWebhookDeliveryStatus(webhookDelivery);

        return new CompletionAckDTO(ackId, deliveryId, completionStatus, completedTimestamp.getTime());
    }
}
