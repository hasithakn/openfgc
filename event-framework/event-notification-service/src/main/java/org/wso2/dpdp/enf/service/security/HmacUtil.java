package org.wso2.dpdp.enf.service.security;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.nio.charset.StandardCharsets;
import java.security.InvalidKeyException;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.HexFormat;

public class HmacUtil {

    private static final String HMAC_SHA256_ALGORITHM = "HmacSHA256";

    private HmacUtil() {
    }

    public static String calculateHmac(String data, String secret) {
        if (data == null || secret == null) {
            return "";
        }
        try {
            SecretKeySpec secretKey = new SecretKeySpec(secret.getBytes(StandardCharsets.UTF_8), HMAC_SHA256_ALGORITHM);
            Mac mac = Mac.getInstance(HMAC_SHA256_ALGORITHM);
            mac.init(secretKey);
            byte[] hmacBytes = mac.doFinal(data.getBytes(StandardCharsets.UTF_8));
            return HexFormat.of().formatHex(hmacBytes);
        } catch (NoSuchAlgorithmException | InvalidKeyException e) {
            throw new RuntimeException("Error calculating HMAC-SHA256 signature", e);
        }
    }

    /**
     * Verifies an HMAC-SHA256 signature using constant-time comparison to prevent timing attacks.
     */
    public static boolean verifySignature(String payload, String secret, String expectedSignature) {
        if (expectedSignature == null || expectedSignature.trim().isEmpty()) {
            return false;
        }
        String computedHex = calculateHmac(payload, secret);
        // Use constant-time comparison to prevent timing-based side-channel attacks
        return MessageDigest.isEqual(
                computedHex.getBytes(StandardCharsets.UTF_8),
                expectedSignature.trim().getBytes(StandardCharsets.UTF_8)
        );
    }
}
