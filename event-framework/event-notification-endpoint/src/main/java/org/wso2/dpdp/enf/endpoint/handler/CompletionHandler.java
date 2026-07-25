package org.wso2.dpdp.enf.endpoint.handler;

import org.wso2.dpdp.enf.endpoint.bean.CompletionResponseBean;
import org.wso2.dpdp.enf.endpoint.bean.CompletionRequestBean;
import org.wso2.dpdp.enf.service.CompletionService;
import org.wso2.dpdp.enf.service.dto.CompletionAckDTO;
import org.wso2.dpdp.enf.service.impl.CompletionServiceImpl;

public class CompletionHandler {

    private final CompletionService completionService;

    public CompletionHandler() {
        this.completionService = new CompletionServiceImpl();
    }

    public CompletionHandler(CompletionService completionService) {
        this.completionService = completionService;
    }

    public CompletionResponseBean submitCompletion(String orgId, String groupId, String deliveryId,
                                                String rawRequestBody, String signature,
                                                CompletionRequestBean request) {
        String status = request != null ? request.getCompletionStatus() : null;
        String evidence = request != null ? request.getCompletionEvidence() : null;
        Long completedAt = request != null ? request.getCompletedAt() : null;

        CompletionAckDTO dto = completionService.submitCompletion(
                orgId, groupId, deliveryId, rawRequestBody, signature, status, evidence, completedAt
        );

        return new CompletionResponseBean(dto.getAckId(), dto.getDeliveryId(), dto.getCompletionStatus(), dto.getCompletedAt());
    }
}
