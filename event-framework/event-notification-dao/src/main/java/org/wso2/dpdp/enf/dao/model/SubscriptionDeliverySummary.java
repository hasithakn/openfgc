package org.wso2.dpdp.enf.dao.model;

import java.sql.Timestamp;

public class SubscriptionDeliverySummary {

    private String deliveryId;
    private String eventId;
    private String subscriptionId;
    private String topicName;
    private String currentStatus;
    private String deliveryMode;
    private Timestamp occurredAt;
    private Timestamp createdAt;
    private String payload;

    public SubscriptionDeliverySummary() {
    }

    public SubscriptionDeliverySummary(String deliveryId, String eventId, String topicName, String currentStatus,
                                       String deliveryMode, Timestamp occurredAt, Timestamp createdAt) {
        this(deliveryId, eventId, null, topicName, currentStatus, deliveryMode, occurredAt, createdAt, null);
    }

    public SubscriptionDeliverySummary(String deliveryId, String eventId, String subscriptionId, String topicName,
                                       String currentStatus, String deliveryMode, Timestamp occurredAt, Timestamp createdAt) {
        this(deliveryId, eventId, subscriptionId, topicName, currentStatus, deliveryMode, occurredAt, createdAt, null);
    }

    public SubscriptionDeliverySummary(String deliveryId, String eventId, String subscriptionId, String topicName,
                                       String currentStatus, String deliveryMode, Timestamp occurredAt, Timestamp createdAt, String payload) {
        this.deliveryId = deliveryId;
        this.eventId = eventId;
        this.subscriptionId = subscriptionId;
        this.topicName = topicName;
        this.currentStatus = currentStatus;
        this.deliveryMode = deliveryMode;
        this.occurredAt = occurredAt;
        this.createdAt = createdAt;
        this.payload = payload;
    }

    public String getDeliveryId() {
        return deliveryId;
    }

    public void setDeliveryId(String deliveryId) {
        this.deliveryId = deliveryId;
    }

    public String getEventId() {
        return eventId;
    }

    public void setEventId(String eventId) {
        this.eventId = eventId;
    }

    public String getSubscriptionId() {
        return subscriptionId;
    }

    public void setSubscriptionId(String subscriptionId) {
        this.subscriptionId = subscriptionId;
    }

    public String getTopicName() {
        return topicName;
    }

    public void setTopicName(String topicName) {
        this.topicName = topicName;
    }

    public String getCurrentStatus() {
        return currentStatus;
    }

    public void setCurrentStatus(String currentStatus) {
        this.currentStatus = currentStatus;
    }

    public String getDeliveryMode() {
        return deliveryMode;
    }

    public void setDeliveryMode(String deliveryMode) {
        this.deliveryMode = deliveryMode;
    }

    public Timestamp getOccurredAt() {
        return occurredAt;
    }

    public void setOccurredAt(Timestamp occurredAt) {
        this.occurredAt = occurredAt;
    }

    public Timestamp getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Timestamp createdAt) {
        this.createdAt = createdAt;
    }

    public String getPayload() {
        return payload;
    }

    public void setPayload(String payload) {
        this.payload = payload;
    }
}
