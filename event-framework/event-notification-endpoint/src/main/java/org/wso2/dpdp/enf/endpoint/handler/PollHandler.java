package org.wso2.dpdp.enf.endpoint.handler;

import org.wso2.dpdp.enf.endpoint.bean.PollDeliveryItemBean;
import org.wso2.dpdp.enf.endpoint.bean.PollErrEntryBean;
import org.wso2.dpdp.enf.endpoint.bean.PollRequestBean;
import org.wso2.dpdp.enf.endpoint.bean.PollResponseBean;
import org.wso2.dpdp.enf.service.PollService;
import org.wso2.dpdp.enf.service.dto.PollDeliveryItemDTO;
import org.wso2.dpdp.enf.service.dto.PollResponseDTO;
import org.wso2.dpdp.enf.service.impl.PollServiceImpl;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class PollHandler {

    private final PollService pollService;

    public PollHandler() {
        this.pollService = new PollServiceImpl();
    }

    public PollHandler(PollService pollService) {
        this.pollService = pollService;
    }

    public PollResponseBean pollEvents(String orgId, String groupId, String rawRequestBody, String signature, PollRequestBean request) {
        List<String> ackList = request != null ? request.getAck() : null;
        Map<String, Map<String, String>> setErrsMap = new HashMap<>();

        if (request != null && request.getSetErrs() != null) {
            for (Map.Entry<String, PollErrEntryBean> entry : request.getSetErrs().entrySet()) {
                Map<String, String> errDetail = new HashMap<>();
                if (entry.getValue() != null) {
                    errDetail.put("err", entry.getValue().getErr());
                    errDetail.put("description", entry.getValue().getDescription());
                }
                setErrsMap.put(entry.getKey(), errDetail);
            }
        }

        Integer maxEvents = request != null ? request.getMaxEvents() : null;
        boolean returnImmediately = request == null || Boolean.TRUE.equals(request.getReturnImmediately());

        PollResponseDTO dto = pollService.pollEvents(orgId, groupId, rawRequestBody, signature, ackList, setErrsMap, maxEvents, returnImmediately);
        List<PollDeliveryItemBean> items = new ArrayList<>();

        if (dto.getEvents() != null) {
            for (PollDeliveryItemDTO itemDTO : dto.getEvents()) {
                items.add(new PollDeliveryItemBean(
                        itemDTO.getDeliveryId(),
                        itemDTO.getEventId(),
                        itemDTO.getTopic(),
                        itemDTO.getPurposes(),
                        itemDTO.getPayload(),
                        itemDTO.getCreatedAt()
                ));
            }
        }

        return new PollResponseBean(items);
    }
}
