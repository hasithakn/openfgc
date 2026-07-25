package org.wso2.dpdp.enf.service.dto;

public class DeliveryConfigDTO {

    private String mode;
    private String callbackUrl;
    private String sharedSecret;

    public DeliveryConfigDTO() {
    }

    public DeliveryConfigDTO(String mode, String callbackUrl, String sharedSecret) {
        this.mode = mode;
        this.callbackUrl = callbackUrl;
        this.sharedSecret = sharedSecret;
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

    public String getSharedSecret() {
        return sharedSecret;
    }

    public void setSharedSecret(String sharedSecret) {
        this.sharedSecret = sharedSecret;
    }
}
