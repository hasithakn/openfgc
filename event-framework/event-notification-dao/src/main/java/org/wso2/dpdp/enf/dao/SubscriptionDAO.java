package org.wso2.dpdp.enf.dao;

import org.wso2.dpdp.enf.dao.model.Subscription;
import java.util.List;
import java.util.Optional;

public interface SubscriptionDAO {
    boolean addSubscription(Subscription subscription);
    Optional<Subscription> getSubscriptionById(String subscriptionId, String orgId);
    List<Subscription> getActiveSubscriptionsForMatching(String orgId, String groupId, String topicId);
    boolean updateSubscriptionStatus(String subscriptionId, String status);
    void addSubscriptionPurposes(String subscriptionId, List<String> purposes);
    List<String> getSubscriptionPurposes(String subscriptionId);
    List<Subscription> listSubscriptions(String orgId, String status, String purposes, String search, int limit, int offset, String sort, int[] totalOut);
    boolean hasPendingDeliveries(String subscriptionId);
    Optional<Subscription> findDuplicateSubscription(String orgId, String groupId, String topicId, String purposeFilterMode, List<String> sortedPurposes);
    Optional<Subscription> getActivePullSubscription(String orgId, String groupId);
}
