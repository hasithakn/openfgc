package org.wso2.dpdp.enf.endpoint.handler;

import org.wso2.dpdp.enf.endpoint.bean.DeliveryConfigBean;
import org.wso2.dpdp.enf.endpoint.bean.DeliveryConfigOutBean;
import org.wso2.dpdp.enf.endpoint.bean.FilterBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionCreateRequestBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionDeliveryAttemptBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionEventHistoryResponseBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionEventItemBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionEventListResponseBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionListResponseBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionResponseBean;
import org.wso2.dpdp.enf.service.SubscriptionService;
import org.wso2.dpdp.enf.service.dto.DeliveryConfigDTO;
import org.wso2.dpdp.enf.service.dto.FilterDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionDeliveryAttemptDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionDeliveryDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionEventHistoryDTO;
import org.wso2.dpdp.enf.service.impl.SubscriptionServiceImpl;

import java.util.ArrayList;
import java.util.List;

public class SubscriptionHandler {

    private final SubscriptionService subscriptionService;

    public SubscriptionHandler() {
        this.subscriptionService = new SubscriptionServiceImpl();
    }

    public SubscriptionHandler(SubscriptionService subscriptionService) {
        this.subscriptionService = subscriptionService;
    }

    public SubscriptionResponseBean createSubscription(String orgId, String groupId, SubscriptionCreateRequestBean request) {
        String topic = request != null ? request.getTopic() : null;
        FilterBean fb = request != null ? request.getFilter() : null;
        DeliveryConfigBean db = request != null ? request.getDelivery() : null;

        FilterDTO filterDTO = fb != null ? new FilterDTO(fb.getType(), fb.getPurposes()) : null;
        DeliveryConfigDTO deliveryDTO = db != null ? new DeliveryConfigDTO(db.getMode(), db.getCallbackUrl(), db.getSharedSecret()) : null;

        SubscriptionDTO dto = subscriptionService.createSubscription(orgId, groupId, topic, filterDTO, deliveryDTO);
        return mapToBean(dto);
    }

    public SubscriptionListResponseBean listSubscriptions(String orgId, String status, String purposes, String search, Integer limit, Integer offset, String sort) {
        int lim = limit != null && limit > 0 ? limit : 20;
        int off = offset != null && offset >= 0 ? offset : 0;
        int[] totalOut = new int[]{0};

        List<SubscriptionDTO> dtos = subscriptionService.listSubscriptions(orgId, status, purposes, search, lim, off, sort, totalOut);
        List<SubscriptionResponseBean> beanList = new ArrayList<>();
        for (SubscriptionDTO dto : dtos) {
            beanList.add(mapToBean(dto));
        }
        return new SubscriptionListResponseBean(beanList, totalOut[0], lim, off, beanList.size());
    }

    public SubscriptionResponseBean getSubscription(String orgId, String subscriptionId) {
        SubscriptionDTO dto = subscriptionService.getSubscription(orgId, subscriptionId);
        return mapToBean(dto);
    }

    public boolean deleteSubscription(String orgId, String subscriptionId) {
        return subscriptionService.deleteSubscription(orgId, subscriptionId);
    }

    public SubscriptionEventListResponseBean listSubscriptionEvents(String orgId, String subscriptionId, Integer limit, Integer offset) {
        int lim = limit != null && limit > 0 ? limit : 20;
        int off = offset != null && offset >= 0 ? offset : 0;
        int[] totalOut = new int[]{0};

        List<SubscriptionDeliveryDTO> dtos = subscriptionService.listSubscriptionEvents(orgId, subscriptionId, lim, off, totalOut);
        List<SubscriptionEventItemBean> itemList = new ArrayList<>();
        for (SubscriptionDeliveryDTO dto : dtos) {
            itemList.add(new SubscriptionEventItemBean(
                    dto.getDeliveryId(),
                    dto.getEventId(),
                    dto.getTopic(),
                    dto.getCurrentStatus(),
                    dto.getDeliveryMode(),
                    dto.getOccurredAt()
            ));
        }
        return new SubscriptionEventListResponseBean(itemList, totalOut[0], lim, off, itemList.size());
    }

    public SubscriptionEventHistoryResponseBean getSubscriptionEventHistory(String orgId, String subscriptionId, String deliveryId) {
        SubscriptionEventHistoryDTO dto = subscriptionService.getSubscriptionEventHistory(orgId, subscriptionId, deliveryId);
        SubscriptionEventHistoryResponseBean bean = new SubscriptionEventHistoryResponseBean(
                dto.getDeliveryId(),
                dto.getEventId(),
                dto.getTopic(),
                dto.getDeliveryMode(),
                dto.getCurrentStatus(),
                dto.getOccurredAt()
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

    private SubscriptionResponseBean mapToBean(SubscriptionDTO dto) {
        FilterBean fb = dto.getFilter() != null ? new FilterBean(dto.getFilter().getType(), dto.getFilter().getPurposes()) : null;
        DeliveryConfigOutBean db = dto.getDelivery() != null ? new DeliveryConfigOutBean(dto.getDelivery().getMode(), dto.getDelivery().getCallbackUrl()) : null;
        SubscriptionResponseBean bean = new SubscriptionResponseBean(
                dto.getSubscriptionId(),
                dto.getTopic(),
                fb,
                db,
                dto.getStatus(),
                dto.getCreatedAt(),
                dto.getUpdatedAt()
        );
        bean.setAlreadyExists(dto.getAlreadyExists());
        bean.setMessage(dto.getMessage());
        return bean;
    }
}
