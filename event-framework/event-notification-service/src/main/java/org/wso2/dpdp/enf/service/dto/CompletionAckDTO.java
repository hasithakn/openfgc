package org.wso2.dpdp.enf.service.dto;

public class CompletionAckDTO {

    private String ackId;
    private String deliveryId;
    private String completionStatus;
    private Long completedAt;

    public CompletionAckDTO() {
    }

    public CompletionAckDTO(String ackId, String deliveryId, String completionStatus, Long completedAt) {
        this.ackId = ackId;
        this.deliveryId = deliveryId;
        this.completionStatus = completionStatus;
        this.completedAt = completedAt;
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

    public String getCompletionStatus() {
        return completionStatus;
    }

    public void setCompletionStatus(String completionStatus) {
        this.completionStatus = completionStatus;
    }

    public Long getCompletedAt() {
        return completedAt;
    }

    public void setCompletedAt(Long completedAt) {
        this.completedAt = completedAt;
    }
}
