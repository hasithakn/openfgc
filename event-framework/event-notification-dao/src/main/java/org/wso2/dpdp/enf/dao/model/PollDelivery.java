package org.wso2.dpdp.enf.dao.model;

import java.sql.Timestamp;

public class PollDelivery {

    private String deliveryId;
    private String subscriptionId;
    private String eventId;
    private String status;
    private Timestamp createdAt;
    private Timestamp completedAt;

    public PollDelivery() {
    }

    public PollDelivery(String deliveryId, String subscriptionId, String eventId,
                        String status, Timestamp createdAt, Timestamp completedAt) {
        this.deliveryId = deliveryId;
        this.subscriptionId = subscriptionId;
        this.eventId = eventId;
        this.status = status;
        this.createdAt = createdAt;
        this.completedAt = completedAt;
    }

    public String getDeliveryId() {
        return deliveryId;
    }

    public void setDeliveryId(String deliveryId) {
        this.deliveryId = deliveryId;
    }

    public String getSubscriptionId() {
        return subscriptionId;
    }

    public void setSubscriptionId(String subscriptionId) {
        this.subscriptionId = subscriptionId;
    }

    public String getEventId() {
        return eventId;
    }

    public void setEventId(String eventId) {
        this.eventId = eventId;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public Timestamp getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Timestamp createdAt) {
        this.createdAt = createdAt;
    }

    public Timestamp getCompletedAt() {
        return completedAt;
    }

    public void setCompletedAt(Timestamp completedAt) {
        this.completedAt = completedAt;
    }
}
