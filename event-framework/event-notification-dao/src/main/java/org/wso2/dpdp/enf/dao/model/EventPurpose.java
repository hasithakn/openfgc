package org.wso2.dpdp.enf.dao.model;

public class EventPurpose {

    private String eventId;
    private String purposeName;

    public EventPurpose() {
    }

    public EventPurpose(String eventId, String purposeName) {
        this.eventId = eventId;
        this.purposeName = purposeName;
    }

    public String getEventId() {
        return eventId;
    }

    public void setEventId(String eventId) {
        this.eventId = eventId;
    }

    public String getPurposeName() {
        return purposeName;
    }

    public void setPurposeName(String purposeName) {
        this.purposeName = purposeName;
    }
}
