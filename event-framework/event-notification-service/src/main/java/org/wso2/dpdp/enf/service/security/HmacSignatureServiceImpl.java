package org.wso2.dpdp.enf.service.security;

public class HmacSignatureServiceImpl implements HmacSignatureService {

    @Override
    public String generateSignature(String payload, String secret) {
        return HmacUtil.calculateHmac(payload, secret);
    }

    @Override
    public boolean verifySignature(String payload, String secret, String signature) {
        return HmacUtil.verifySignature(payload, secret, signature);
    }
}
