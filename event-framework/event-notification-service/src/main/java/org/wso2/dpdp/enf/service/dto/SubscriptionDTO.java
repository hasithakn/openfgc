package org.wso2.dpdp.enf.service.dto;

public class SubscriptionDTO {

    private String subscriptionId;
    private String topic;
    private FilterDTO filter;
    private DeliveryConfigDTO delivery;
    private String status;
    private Long createdAt;
    private Long updatedAt;
    private Boolean alreadyExists;
    private String message;

    public SubscriptionDTO() {
    }

    public SubscriptionDTO(String subscriptionId, String topic, FilterDTO filter, DeliveryConfigDTO delivery,
                           String status, Long createdAt, Long updatedAt) {
        this.subscriptionId = subscriptionId;
        this.topic = topic;
        this.filter = filter;
        this.delivery = delivery;
        this.status = status;
        this.createdAt = createdAt;
        this.updatedAt = updatedAt;
    }

    public String getSubscriptionId() {
        return subscriptionId;
    }

    public void setSubscriptionId(String subscriptionId) {
        this.subscriptionId = subscriptionId;
    }

    public String getTopic() {
        return topic;
    }

    public void setTopic(String topic) {
        this.topic = topic;
    }

    public FilterDTO getFilter() {
        return filter;
    }

    public void setFilter(FilterDTO filter) {
        this.filter = filter;
    }

    public DeliveryConfigDTO getDelivery() {
        return delivery;
    }

    public void setDelivery(DeliveryConfigDTO delivery) {
        this.delivery = delivery;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public Long getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Long createdAt) {
        this.createdAt = createdAt;
    }

    public Long getUpdatedAt() {
        return updatedAt;
    }

    public void setUpdatedAt(Long updatedAt) {
        this.updatedAt = updatedAt;
    }

    public Boolean getAlreadyExists() {
        return alreadyExists;
    }

    public void setAlreadyExists(Boolean alreadyExists) {
        this.alreadyExists = alreadyExists;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }
}
