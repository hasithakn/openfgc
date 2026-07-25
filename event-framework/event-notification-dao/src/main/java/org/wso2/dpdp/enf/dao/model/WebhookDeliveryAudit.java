package org.wso2.dpdp.enf.dao.model;

import java.sql.Timestamp;

public class WebhookDeliveryAudit {

    private String auditId;
    private String eventId;
    private String deliveryId;
    private String orgId;
    private String responseCode;
    private Timestamp createdAt;
    private Timestamp attemptAt;

    public WebhookDeliveryAudit() {
    }

    public WebhookDeliveryAudit(String auditId, String eventId, String deliveryId, String orgId,
                                String responseCode, Timestamp createdAt, Timestamp attemptAt) {
        this.auditId = auditId;
        this.eventId = eventId;
        this.deliveryId = deliveryId;
        this.orgId = orgId;
        this.responseCode = responseCode;
        this.createdAt = createdAt;
        this.attemptAt = attemptAt;
    }

    public String getAuditId() {
        return auditId;
    }

    public void setAuditId(String auditId) {
        this.auditId = auditId;
    }

    public String getEventId() {
        return eventId;
    }

    public void setEventId(String eventId) {
        this.eventId = eventId;
    }

    public String getDeliveryId() {
        return deliveryId;
    }

    public void setDeliveryId(String deliveryId) {
        this.deliveryId = deliveryId;
    }

    public String getOrgId() {
        return orgId;
    }

    public void setOrgId(String orgId) {
        this.orgId = orgId;
    }

    public String getResponseCode() {
        return responseCode;
    }

    public void setResponseCode(String responseCode) {
        this.responseCode = responseCode;
    }

    public Timestamp getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Timestamp createdAt) {
        this.createdAt = createdAt;
    }

    public Timestamp getAttemptAt() {
        return attemptAt;
    }

    public void setAttemptAt(Timestamp attemptAt) {
        this.attemptAt = attemptAt;
    }
}
