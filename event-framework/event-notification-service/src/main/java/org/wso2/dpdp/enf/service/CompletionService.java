package org.wso2.dpdp.enf.service;

import org.wso2.dpdp.enf.service.dto.CompletionAckDTO;

public interface CompletionService {
    CompletionAckDTO submitCompletion(String orgId, String groupId, String deliveryId,
                                      String rawRequestBody, String signature,
                                      String completionStatus, String completionEvidence, Long completedAt);
}
