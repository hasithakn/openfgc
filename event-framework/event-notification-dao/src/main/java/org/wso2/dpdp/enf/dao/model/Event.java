package org.wso2.dpdp.enf.dao.model;

import java.sql.Timestamp;
import java.util.List;

public class Event {

    private String eventId;
    private String orgId;
    private String groupId;
    private String topicId;
    private String payload;
    private Timestamp createdAt;
    private List<String> purposes;

    public Event() {
    }

    public Event(String eventId, String orgId, String groupId, String topicId, String payload, Timestamp createdAt) {
        this.eventId = eventId;
        this.orgId = orgId;
        this.groupId = groupId;
        this.topicId = topicId;
        this.payload = payload;
        this.createdAt = createdAt;
    }

    public String getEventId() {
        return eventId;
    }

    public void setEventId(String eventId) {
        this.eventId = eventId;
    }

    public String getOrgId() {
        return orgId;
    }

    public void setOrgId(String orgId) {
        this.orgId = orgId;
    }

    public String getGroupId() {
        return groupId;
    }

    public void setGroupId(String groupId) {
        this.groupId = groupId;
    }

    public String getTopicId() {
        return topicId;
    }

    public void setTopicId(String topicId) {
        this.topicId = topicId;
    }

    public String getPayload() {
        return payload;
    }

    public void setPayload(String payload) {
        this.payload = payload;
    }

    public Timestamp getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Timestamp createdAt) {
        this.createdAt = createdAt;
    }

    public List<String> getPurposes() {
        return purposes;
    }

    public void setPurposes(List<String> purposes) {
        this.purposes = purposes;
    }
}
