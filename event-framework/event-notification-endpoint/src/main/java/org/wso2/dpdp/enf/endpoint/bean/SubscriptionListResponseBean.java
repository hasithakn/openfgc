package org.wso2.dpdp.enf.endpoint.bean;

import java.util.List;

public class SubscriptionListResponseBean {

    private List<SubscriptionResponseBean> subscriptions;
    private int total;
    private int limit;
    private int offset;
    private int count;

    public SubscriptionListResponseBean() {
    }

    public SubscriptionListResponseBean(List<SubscriptionResponseBean> subscriptions, int total, int limit, int offset, int count) {
        this.subscriptions = subscriptions;
        this.total = total;
        this.limit = limit;
        this.offset = offset;
        this.count = count;
    }

    public List<SubscriptionResponseBean> getSubscriptions() {
        return subscriptions;
    }

    public void setSubscriptions(List<SubscriptionResponseBean> subscriptions) {
        this.subscriptions = subscriptions;
    }

    public int getTotal() {
        return total;
    }

    public void setTotal(int total) {
        this.total = total;
    }

    public int getLimit() {
        return limit;
    }

    public void setLimit(int limit) {
        this.limit = limit;
    }

    public int getOffset() {
        return offset;
    }

    public void setOffset(int offset) {
        this.offset = offset;
    }

    public int getCount() {
        return count;
    }

    public void setCount(int count) {
        this.count = count;
    }
}
