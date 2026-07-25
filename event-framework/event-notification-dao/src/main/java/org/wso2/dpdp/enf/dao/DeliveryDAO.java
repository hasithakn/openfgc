package org.wso2.dpdp.enf.dao;

import org.wso2.dpdp.enf.dao.model.WebhookDelivery;
import org.wso2.dpdp.enf.dao.model.WebhookDeliveryAudit;
import org.wso2.dpdp.enf.dao.model.PollDelivery;
import org.wso2.dpdp.enf.dao.model.SubscriptionDeliverySummary;
import java.util.List;
import java.util.Optional;

public interface DeliveryDAO {
    boolean addWebhookDelivery(WebhookDelivery delivery);
    Optional<WebhookDelivery> getWebhookDeliveryById(String deliveryId, String orgId);
    List<WebhookDelivery> getPendingWebhookDeliveries(int limit);
    boolean updateWebhookDeliveryStatus(WebhookDelivery delivery);
    
    boolean addWebhookDeliveryAudit(WebhookDeliveryAudit audit);
    List<WebhookDeliveryAudit> getWebhookDeliveryAudits(String deliveryId);

    boolean addPollDelivery(PollDelivery delivery);
    Optional<PollDelivery> getPollDeliveryById(String deliveryId, String orgId);
    List<PollDelivery> getPendingPollDeliveries(String orgId, String groupId, int limit);
    boolean updatePollDeliveryStatus(String deliveryId, String status);
    boolean isDeliveryExistsForOrg(String deliveryId, String orgId);

    List<SubscriptionDeliverySummary> listSubscriptionDeliveries(String subscriptionId, int limit, int offset, int[] totalOut);
    Optional<SubscriptionDeliverySummary> getSubscriptionDeliveryById(String subscriptionId, String deliveryId);

    List<SubscriptionDeliverySummary> listOrgDeliveries(String orgId, String statusFilter, String subscriptionIdFilter, String purposesFilter, String search, int limit, int offset, int[] totalOut);
    Optional<SubscriptionDeliverySummary> getOrgDeliveryById(String orgId, String deliveryId);
}
