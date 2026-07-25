package org.wso2.dpdp.enf.dao.model;

public class SubscriptionPurpose {

    private String subscriptionId;
    private String purposeName;

    public SubscriptionPurpose() {
    }

    public SubscriptionPurpose(String subscriptionId, String purposeName) {
        this.subscriptionId = subscriptionId;
        this.purposeName = purposeName;
    }

    public String getSubscriptionId() {
        return subscriptionId;
    }

    public void setSubscriptionId(String subscriptionId) {
        this.subscriptionId = subscriptionId;
    }

    public String getPurposeName() {
        return purposeName;
    }

    public void setPurposeName(String purposeName) {
        this.purposeName = purposeName;
    }
}
