package org.wso2.dpdp.enf.service.dto;

import java.util.List;
import java.util.Map;

public class EventDTO {

    private String eventId;
    private String topic;
    private List<String> purposes;
    private Map<String, Object> payload;
    private long occurredAt;
    private long createdAt;

    public EventDTO() {
    }

    public EventDTO(String eventId, String topic, List<String> purposes, Map<String, Object> payload,
                    long occurredAt, long createdAt) {
        this.eventId = eventId;
        this.topic = topic;
        this.purposes = purposes;
        this.payload = payload;
        this.occurredAt = occurredAt;
        this.createdAt = createdAt;
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

    public long getOccurredAt() {
        return occurredAt;
    }

    public void setOccurredAt(long occurredAt) {
        this.occurredAt = occurredAt;
    }

    public long getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(long createdAt) {
        this.createdAt = createdAt;
    }
}
