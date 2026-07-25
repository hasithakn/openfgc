package org.wso2.dpdp.enf.endpoint.bean;

import com.fasterxml.jackson.annotation.JsonInclude;

@JsonInclude(JsonInclude.Include.NON_NULL)
public class SubscriptionResponseBean {

    private String subscriptionId;
    private String topic;
    private FilterBean filter;
    private DeliveryConfigOutBean delivery;
    private String status;
    private Long createdAt;
    private Long updatedAt;
    private Boolean alreadyExists;
    private String message;

    public SubscriptionResponseBean() {
    }

    public SubscriptionResponseBean(String subscriptionId, String topic, FilterBean filter,
                                    DeliveryConfigOutBean delivery, String status,
                                    Long createdAt, Long updatedAt) {
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

    public FilterBean getFilter() {
        return filter;
    }

    public void setFilter(FilterBean filter) {
        this.filter = filter;
    }

    public DeliveryConfigOutBean getDelivery() {
        return delivery;
    }

    public void setDelivery(DeliveryConfigOutBean delivery) {
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
