package org.wso2.dpdp.enf.endpoint.bean;

public class SubscriptionEventItemBean {

    private String deliveryId;
    private String eventId;
    private String topic;
    private String currentStatus;
    private String deliveryMode;
    private long occurredAt;
    private Object payload;

    public SubscriptionEventItemBean() {
    }

    public SubscriptionEventItemBean(String deliveryId, String eventId, String topic, String currentStatus,
                                     String deliveryMode, long occurredAt) {
        this(deliveryId, eventId, topic, currentStatus, deliveryMode, occurredAt, null);
    }

    public SubscriptionEventItemBean(String deliveryId, String eventId, String topic, String currentStatus,
                                     String deliveryMode, long occurredAt, Object payload) {
        this.deliveryId = deliveryId;
        this.eventId = eventId;
        this.topic = topic;
        this.currentStatus = currentStatus;
        this.deliveryMode = deliveryMode;
        this.occurredAt = occurredAt;
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

    public String getTopic() {
        return topic;
    }

    public void setTopic(String topic) {
        this.topic = topic;
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

    public long getOccurredAt() {
        return occurredAt;
    }

    public void setOccurredAt(long occurredAt) {
        this.occurredAt = occurredAt;
    }

    public Object getPayload() {
        return payload;
    }

    public void setPayload(Object payload) {
        this.payload = payload;
    }
}
