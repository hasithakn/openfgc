package org.wso2.dpdp.enf.endpoint.bean;

public class SubscriptionCreateRequestBean {

    private String topic;
    private FilterBean filter;
    private DeliveryConfigBean delivery;

    public SubscriptionCreateRequestBean() {
    }

    public SubscriptionCreateRequestBean(String topic, FilterBean filter, DeliveryConfigBean delivery) {
        this.topic = topic;
        this.filter = filter;
        this.delivery = delivery;
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

    public DeliveryConfigBean getDelivery() {
        return delivery;
    }

    public void setDelivery(DeliveryConfigBean delivery) {
        this.delivery = delivery;
    }
}
