package org.wso2.dpdp.enf.service;

import org.wso2.dpdp.enf.service.dto.EventDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionDeliveryDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionEventHistoryDTO;
import java.util.List;
import java.util.Map;

public interface EventPublishService {
    EventDTO publishEvent(String orgId, String groupId, String topicName, List<String> purposes, Map<String, Object> payload);
    List<SubscriptionDeliveryDTO> listOrgEvents(String orgId, String statusFilter, String subscriptionIdFilter, String purposesFilter, String search, int limit, int offset, int[] totalOut);
    SubscriptionEventHistoryDTO getOrgEventHistory(String orgId, String deliveryId);
}
