package org.wso2.dpdp.enf.endpoint.bean;

import java.util.List;

public class PollResponseBean {

    private List<PollDeliveryItemBean> events;

    public PollResponseBean() {
    }

    public PollResponseBean(List<PollDeliveryItemBean> events) {
        this.events = events;
    }

    public List<PollDeliveryItemBean> getEvents() {
        return events;
    }

    public void setEvents(List<PollDeliveryItemBean> events) {
        this.events = events;
    }
}
