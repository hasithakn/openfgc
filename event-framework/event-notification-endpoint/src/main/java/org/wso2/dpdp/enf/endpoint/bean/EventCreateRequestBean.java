package org.wso2.dpdp.enf.endpoint.bean;

import java.util.List;
import java.util.Map;

public class EventCreateRequestBean {

    private String topic;
    private List<String> purposes;
    private Map<String, Object> payload;

    public EventCreateRequestBean() {
    }

    public EventCreateRequestBean(String topic, List<String> purposes, Map<String, Object> payload) {
        this.topic = topic;
        this.purposes = purposes;
        this.payload = payload;
    }

    public String getTopic() {
        return topic;
    }

    public void setTopic(String topic) {
        this.topic = topic;
    }

    public List<String> getPurposes() {
        return purposes;
    }

    public void setPurposes(List<String> purposes) {
        this.purposes = purposes;
    }

    public Map<String, Object> getPayload() {
        return payload;
    }

    public void setPayload(Map<String, Object> payload) {
        this.payload = payload;
    }
}
