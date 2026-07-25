package org.wso2.dpdp.enf.service;

import org.wso2.dpdp.enf.service.dto.PollResponseDTO;
import java.util.List;
import java.util.Map;

public interface PollService {
    PollResponseDTO pollEvents(String orgId, String groupId, String rawRequestBody, String signature,
                              List<String> ackDeliveryIds, Map<String, Map<String, String>> setErrs,
                              Integer maxEvents, boolean returnImmediately);
}
