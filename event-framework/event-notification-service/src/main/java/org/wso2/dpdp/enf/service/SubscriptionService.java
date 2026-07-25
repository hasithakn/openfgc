package org.wso2.dpdp.enf.service;

import org.wso2.dpdp.enf.service.dto.SubscriptionDTO;
import org.wso2.dpdp.enf.service.dto.FilterDTO;
import org.wso2.dpdp.enf.service.dto.DeliveryConfigDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionDeliveryDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionEventHistoryDTO;
import java.util.List;

public interface SubscriptionService {
    SubscriptionDTO createSubscription(String orgId, String groupId, String topicName, FilterDTO filter, DeliveryConfigDTO delivery);
    List<SubscriptionDTO> listSubscriptions(String orgId, String status, String purposes, String search, int limit, int offset, String sort, int[] totalOut);
    SubscriptionDTO getSubscription(String orgId, String subscriptionId);
    boolean deleteSubscription(String orgId, String subscriptionId);

    List<SubscriptionDeliveryDTO> listSubscriptionEvents(String orgId, String subscriptionId, int limit, int offset, int[] totalOut);
    SubscriptionEventHistoryDTO getSubscriptionEventHistory(String orgId, String subscriptionId, String deliveryId);
}
