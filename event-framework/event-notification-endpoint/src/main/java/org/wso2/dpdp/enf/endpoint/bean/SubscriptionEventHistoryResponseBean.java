package org.wso2.dpdp.enf.endpoint.bean;

import com.fasterxml.jackson.annotation.JsonInclude;
import java.util.ArrayList;
import java.util.List;

@JsonInclude(JsonInclude.Include.NON_NULL)
public class SubscriptionEventHistoryResponseBean {

    private String deliveryId;
    private String eventId;
    private String topic;
    private String deliveryMode;
    private String currentStatus;
    private long occurredAt;
    private Long nextRetryAt;
    private String completionStatus;
    private String completionEvidence;
    private Object payload;
    private List<SubscriptionDeliveryAttemptBean> history = new ArrayList<>();

    public SubscriptionEventHistoryResponseBean() {
    }

    public SubscriptionEventHistoryResponseBean(String deliveryId, String eventId, String topic,
                                                String deliveryMode, String currentStatus, long occurredAt) {
        this(deliveryId, eventId, topic, deliveryMode, currentStatus, occurredAt, null);
    }

    public SubscriptionEventHistoryResponseBean(String deliveryId, String eventId, String topic,
                                                String deliveryMode, String currentStatus, long occurredAt, Object payload) {
        this.deliveryId = deliveryId;
        this.eventId = eventId;
        this.topic = topic;
        this.deliveryMode = deliveryMode;
        this.currentStatus = currentStatus;
        this.occurredAt = occurredAt;
        this.payload = payload;
    }

    public String getDeliveryId() {
        return deliveryId;
    }

    public void setDeliveryId(String deliveryId) {
        this.deliveryId = deliveryId;
    }

    public String getEventId() {
        return eventId;
    }

    public void setEventId(String eventId) {
        this.eventId = eventId;
    }

    public String getTopic() {
        return topic;
    }

    public void setTopic(String topic) {
        this.topic = topic;
    }

    public String getDeliveryMode() {
        return deliveryMode;
    }

    public void setDeliveryMode(String deliveryMode) {
        this.deliveryMode = deliveryMode;
    }

    public String getCurrentStatus() {
        return currentStatus;
    }

    public void setCurrentStatus(String currentStatus) {
        this.currentStatus = currentStatus;
    }

    public long getOccurredAt() {
        return occurredAt;
    }

    public void setOccurredAt(long occurredAt) {
        this.occurredAt = occurredAt;
    }

    public Long getNextRetryAt() {
        return nextRetryAt;
    }

    public void setNextRetryAt(Long nextRetryAt) {
        this.nextRetryAt = nextRetryAt;
    }

    public String getCompletionStatus() {
        return completionStatus;
    }

    public void setCompletionStatus(String completionStatus) {
        this.completionStatus = completionStatus;
    }

    public String getCompletionEvidence() {
        return completionEvidence;
    }

    public void setCompletionEvidence(String completionEvidence) {
        this.completionEvidence = completionEvidence;
    }

    public List<SubscriptionDeliveryAttemptBean> getHistory() {
        return history;
    }

    public void setHistory(List<SubscriptionDeliveryAttemptBean> history) {
        this.history = history;
    }

    public Object getPayload() {
        return payload;
    }

    public void setPayload(Object payload) {
        this.payload = payload;
    }
}
