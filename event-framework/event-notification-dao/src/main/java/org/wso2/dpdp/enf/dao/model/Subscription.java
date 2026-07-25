package org.wso2.dpdp.enf.dao.model;

import java.sql.Timestamp;
import java.util.List;

public class Subscription {

    private String subscriptionId;
    private String orgId;
    private String groupId;
    private String topicId;
    private String status;
    private String purposeFilterMode;
    private String callbackUrl;
    private String sharedSecret;
    private Timestamp createdAt;
    private Timestamp updatedAt;
    private String deliveryMode;
    private List<String> purposes;

    public Subscription() {
    }

    public Subscription(String subscriptionId, String orgId, String groupId, String topicId, String status,
                        String purposeFilterMode, String callbackUrl, String sharedSecret,
                        Timestamp createdAt, Timestamp updatedAt, String deliveryMode) {
        this.subscriptionId = subscriptionId;
        this.orgId = orgId;
        this.groupId = groupId;
        this.topicId = topicId;
        this.status = status;
        this.purposeFilterMode = purposeFilterMode;
        this.callbackUrl = callbackUrl;
        this.sharedSecret = sharedSecret;
        this.createdAt = createdAt;
        this.updatedAt = updatedAt;
        this.deliveryMode = deliveryMode;
    }

    public String getSubscriptionId() {
        return subscriptionId;
    }

    public void setSubscriptionId(String subscriptionId) {
        this.subscriptionId = subscriptionId;
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

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public String getPurposeFilterMode() {
        return purposeFilterMode;
    }

    public void setPurposeFilterMode(String purposeFilterMode) {
        this.purposeFilterMode = purposeFilterMode;
    }

    public String getCallbackUrl() {
        return callbackUrl;
    }

    public void setCallbackUrl(String callbackUrl) {
        this.callbackUrl = callbackUrl;
    }

    public String getSharedSecret() {
        return sharedSecret;
    }

    public void setSharedSecret(String sharedSecret) {
        this.sharedSecret = sharedSecret;
    }

    public Timestamp getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Timestamp createdAt) {
        this.createdAt = createdAt;
    }

    public Timestamp getUpdatedAt() {
        return updatedAt;
    }

    public void setUpdatedAt(Timestamp updatedAt) {
        this.updatedAt = updatedAt;
    }

    public String getDeliveryMode() {
        return deliveryMode;
    }

    public void setDeliveryMode(String deliveryMode) {
        this.deliveryMode = deliveryMode;
    }

    public List<String> getPurposes() {
        return purposes;
    }

    public void setPurposes(List<String> purposes) {
        this.purposes = purposes;
    }
}
