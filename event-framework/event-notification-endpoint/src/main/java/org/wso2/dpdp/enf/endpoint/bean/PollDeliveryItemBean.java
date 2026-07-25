package org.wso2.dpdp.enf.endpoint.bean;

import java.util.List;
import java.util.Map;

public class PollDeliveryItemBean {

    private String deliveryId;
    private String eventId;
    private String topic;
    private List<String> purposes;
    private Map<String, Object> payload;
    private long createdAt;

    public PollDeliveryItemBean() {
    }

    public PollDeliveryItemBean(String deliveryId, String eventId, String topic, List<String> purposes,
                                Map<String, Object> payload, long createdAt) {
        this.deliveryId = deliveryId;
        this.eventId = eventId;
        this.topic = topic;
        this.purposes = purposes;
        this.payload = payload;
        this.createdAt = createdAt;
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

    public List<String> getPurposes() {
        return purposes;
    }

    public void setPurposes(List<String> purposes) {
        this.purposes = purposes;
    }

    public Map<String, Object> getPayload() {
        return payload;
    }

    public void setPayload(Map<String, Object> payload) {
        this.payload = payload;
    }

    public long getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(long createdAt) {
        this.createdAt = createdAt;
    }
}
