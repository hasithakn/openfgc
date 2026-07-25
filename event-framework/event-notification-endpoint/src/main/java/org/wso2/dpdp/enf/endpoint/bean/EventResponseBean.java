package org.wso2.dpdp.enf.endpoint.bean;

import java.util.List;

public class EventResponseBean {

    private String eventId;
    private String topic;
    private List<String> purposes;
    private long occurredAt;
    private long createdAt;

    public EventResponseBean() {
    }

    public EventResponseBean(String eventId, String topic, List<String> purposes, long occurredAt, long createdAt) {
        this.eventId = eventId;
        this.topic = topic;
        this.purposes = purposes;
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
