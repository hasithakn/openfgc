package org.wso2.dpdp.enf.endpoint.bean;

import java.util.List;
import java.util.Map;

public class PollRequestBean {

    private List<String> ack;
    private Map<String, PollErrEntryBean> setErrs;
    private Integer maxEvents;
    private Boolean returnImmediately;

    public PollRequestBean() {
    }

    public PollRequestBean(List<String> ack, Map<String, PollErrEntryBean> setErrs, Integer maxEvents, Boolean returnImmediately) {
        this.ack = ack;
        this.setErrs = setErrs;
        this.maxEvents = maxEvents;
        this.returnImmediately = returnImmediately;
    }

    public List<String> getAck() {
        return ack;
    }

    public void setAck(List<String> ack) {
        this.ack = ack;
    }

    public Map<String, PollErrEntryBean> getSetErrs() {
        return setErrs;
    }

    public void setSetErrs(Map<String, PollErrEntryBean> setErrs) {
        this.setErrs = setErrs;
    }

    public Integer getMaxEvents() {
        return maxEvents;
    }

    public void setMaxEvents(Integer maxEvents) {
        this.maxEvents = maxEvents;
    }

    public Boolean getReturnImmediately() {
        return returnImmediately;
    }

    public void setReturnImmediately(Boolean returnImmediately) {
        this.returnImmediately = returnImmediately;
    }
}
