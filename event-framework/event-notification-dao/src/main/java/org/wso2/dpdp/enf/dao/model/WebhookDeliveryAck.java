package org.wso2.dpdp.enf.dao.model;

import java.sql.Timestamp;

public class WebhookDeliveryAck {

    private String ackId;
    private String deliveryId;
    private Timestamp completedAt;
    private String completionStatus;
    private String completionEvidence;

    public WebhookDeliveryAck() {
    }

    public WebhookDeliveryAck(String ackId, String deliveryId, Timestamp completedAt,
                              String completionStatus, String completionEvidence) {
        this.ackId = ackId;
        this.deliveryId = deliveryId;
        this.completedAt = completedAt;
        this.completionStatus = completionStatus;
        this.completionEvidence = completionEvidence;
    }

    public String getAckId() {
        return ackId;
    }

    public void setAckId(String ackId) {
        this.ackId = ackId;
    }

    public String getDeliveryId() {
        return deliveryId;
    }

    public void setDeliveryId(String deliveryId) {
        this.deliveryId = deliveryId;
    }

    public Timestamp getCompletedAt() {
        return completedAt;
    }

    public void setCompletedAt(Timestamp completedAt) {
        this.completedAt = completedAt;
    }

    public String getCompletionStatus() {
        return completionStatus;
    }

    public void setCompletionStatus(String completionStatus) {
        this.completionStatus = completionStatus;
    }

    public String getCompletionEvidence() {
        return completionEvidence;
    }

    public void setCompletionEvidence(String completionEvidence) {
        this.completionEvidence = completionEvidence;
    }
}
