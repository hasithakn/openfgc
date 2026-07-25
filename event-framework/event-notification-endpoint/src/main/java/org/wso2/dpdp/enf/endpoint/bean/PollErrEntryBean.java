package org.wso2.dpdp.enf.endpoint.bean;

public class PollErrEntryBean {

    private String err;
    private String description;

    public PollErrEntryBean() {
    }

    public PollErrEntryBean(String err, String description) {
        this.err = err;
        this.description = description;
    }

    public String getErr() {
        return err;
    }

    public void setErr(String err) {
        this.err = err;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }
}
