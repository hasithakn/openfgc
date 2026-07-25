package org.wso2.dpdp.enf.endpoint.bean;

public class DeliveryConfigOutBean {

    private String mode;
    private String callbackUrl;

    public DeliveryConfigOutBean() {
    }

    public DeliveryConfigOutBean(String mode, String callbackUrl) {
        this.mode = mode;
        this.callbackUrl = callbackUrl;
    }

    public String getMode() {
        return mode;
    }

    public void setMode(String mode) {
        this.mode = mode;
    }

    public String getCallbackUrl() {
        return callbackUrl;
    }

    public void setCallbackUrl(String callbackUrl) {
        this.callbackUrl = callbackUrl;
    }
}
