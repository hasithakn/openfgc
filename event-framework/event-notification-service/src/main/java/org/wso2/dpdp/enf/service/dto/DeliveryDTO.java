package org.wso2.dpdp.enf.service.dto;

/**
 * @deprecated Use {@link DeliveryConfigDTO} instead.
 */
@Deprecated
public class DeliveryDTO extends DeliveryConfigDTO {

    public DeliveryDTO() {
        super();
    }

    public DeliveryDTO(String mode, String callbackUrl, String sharedSecret) {
        super(mode, callbackUrl, sharedSecret);
    }
}
