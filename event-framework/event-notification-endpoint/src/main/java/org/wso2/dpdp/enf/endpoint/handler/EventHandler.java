package org.wso2.dpdp.enf.endpoint.handler;

import org.wso2.dpdp.enf.endpoint.bean.EventCreateRequestBean;
import org.wso2.dpdp.enf.endpoint.bean.EventResponseBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionDeliveryAttemptBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionEventHistoryResponseBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionEventItemBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionEventListResponseBean;
import org.wso2.dpdp.enf.service.EventPublishService;
import org.wso2.dpdp.enf.service.dto.EventDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionDeliveryAttemptDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionDeliveryDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionEventHistoryDTO;
import org.wso2.dpdp.enf.service.impl.EventPublishServiceImpl;

import java.util.ArrayList;
import java.util.List;

public class EventHandler {

    private final EventPublishService eventPublishService;

    public EventHandler() {
        this.eventPublishService = new EventPublishServiceImpl();
    }

    public EventHandler(EventPublishService eventPublishService) {
        this.eventPublishService = eventPublishService;
    }

    public EventResponseBean publishEvent(String orgId, String groupId, EventCreateRequestBean request) {
        String topic = request != null ? request.getTopic() : null;
        EventDTO dto = eventPublishService.publishEvent(
                orgId,
                groupId,
                topic,
                request != null ? request.getPurposes() : null,
                request != null ? request.getPayload() : null
        );
        return new EventResponseBean(
                dto.getEventId(),
                dto.getTopic(),
                dto.getPurposes(),
                dto.getOccurredAt(),
                dto.getCreatedAt()
        );
    }

    public SubscriptionEventListResponseBean listOrgEvents(String orgId, String status, String subscriptionId, String purposes, String search, Integer limit, Integer offset) {
        int lim = limit != null && limit > 0 ? limit : 20;
        int off = offset != null && offset >= 0 ? offset : 0;
        int[] totalOut = new int[]{0};

        List<SubscriptionDeliveryDTO> dtos = eventPublishService.listOrgEvents(orgId, status, subscriptionId, purposes, search, lim, off, totalOut);
        List<SubscriptionEventItemBean> itemList = new ArrayList<>();
        for (SubscriptionDeliveryDTO dto : dtos) {
            itemList.add(new SubscriptionEventItemBean(
                    dto.getDeliveryId(),
                    dto.getEventId(),
                    dto.getTopic(),
                    dto.getCurrentStatus(),
                    dto.getDeliveryMode(),
                    dto.getOccurredAt(),
                    dto.getPayload()
            ));
        }
        return new SubscriptionEventListResponseBean(itemList, totalOut[0], lim, off, itemList.size());
    }

    public SubscriptionEventHistoryResponseBean getOrgEventHistory(String orgId, String deliveryId) {
        SubscriptionEventHistoryDTO dto = eventPublishService.getOrgEventHistory(orgId, deliveryId);
        SubscriptionEventHistoryResponseBean bean = new SubscriptionEventHistoryResponseBean(
                dto.getDeliveryId(),
                dto.getEventId(),
                dto.getTopic(),
                dto.getDeliveryMode(),
                dto.getCurrentStatus(),
                dto.getOccurredAt(),
                dto.getPayload()
        );
        bean.setNextRetryAt(dto.getNextRetryAt());
        bean.setCompletionStatus(dto.getCompletionStatus());
        bean.setCompletionEvidence(dto.getCompletionEvidence());

        List<SubscriptionDeliveryAttemptBean> attempts = new ArrayList<>();
        if (dto.getHistory() != null) {
            for (SubscriptionDeliveryAttemptDTO att : dto.getHistory()) {
                attempts.add(new SubscriptionDeliveryAttemptBean(
                        att.getAttempt(),
                        att.getStatus(),
                        att.getTimestamp(),
                        att.getHttpStatus(),
                        att.getError()
                ));
            }
        }
        bean.setHistory(attempts);
        return bean;
    }
}
