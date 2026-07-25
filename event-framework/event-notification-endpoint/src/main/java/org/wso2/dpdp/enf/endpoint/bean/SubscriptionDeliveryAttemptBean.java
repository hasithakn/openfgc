package org.wso2.dpdp.enf.endpoint.bean;

import com.fasterxml.jackson.annotation.JsonInclude;

@JsonInclude(JsonInclude.Include.NON_NULL)
public class SubscriptionDeliveryAttemptBean {

    private int attempt;
    private String status;
    private long timestamp;
    private Integer httpStatus;
    private String error;

    public SubscriptionDeliveryAttemptBean() {
    }

    public SubscriptionDeliveryAttemptBean(int attempt, String status, long timestamp, Integer httpStatus, String error) {
        this.attempt = attempt;
        this.status = status;
        this.timestamp = timestamp;
        this.httpStatus = httpStatus;
        this.error = error;
    }

    public int getAttempt() {
        return attempt;
    }

    public void setAttempt(int attempt) {
        this.attempt = attempt;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public long getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(long timestamp) {
        this.timestamp = timestamp;
    }

    public Integer getHttpStatus() {
        return httpStatus;
    }

    public void setHttpStatus(Integer httpStatus) {
        this.httpStatus = httpStatus;
    }

    public String getError() {
        return error;
    }

    public void setError(String error) {
        this.error = error;
    }
}
