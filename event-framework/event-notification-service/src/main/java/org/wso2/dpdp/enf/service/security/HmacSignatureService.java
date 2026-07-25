package org.wso2.dpdp.enf.service.security;

public interface HmacSignatureService {
    String generateSignature(String payload, String secret);
    boolean verifySignature(String payload, String secret, String signature);
}
