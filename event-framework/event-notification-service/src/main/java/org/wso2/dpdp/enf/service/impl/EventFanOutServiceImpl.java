package org.wso2.dpdp.enf.service.impl;

import org.wso2.dpdp.enf.dao.DeliveryDAO;
import org.wso2.dpdp.enf.dao.SubscriptionDAO;
import org.wso2.dpdp.enf.dao.impl.DeliveryDAOImpl;
import org.wso2.dpdp.enf.dao.impl.SubscriptionDAOImpl;
import org.wso2.dpdp.enf.dao.model.Event;
import org.wso2.dpdp.enf.dao.model.PollDelivery;
import org.wso2.dpdp.enf.dao.model.Subscription;
import org.wso2.dpdp.enf.dao.model.WebhookDelivery;
import org.wso2.dpdp.enf.service.EventFanOutService;
import org.wso2.dpdp.enf.service.matching.FilterMatcher;

import java.sql.Timestamp;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.UUID;
import java.util.logging.Logger;

public class EventFanOutServiceImpl implements EventFanOutService {

    private static final Logger LOGGER = Logger.getLogger(EventFanOutServiceImpl.class.getName());

    private final SubscriptionDAO subscriptionDAO;
    private final DeliveryDAO deliveryDAO;

    public EventFanOutServiceImpl() {
        this.subscriptionDAO = new SubscriptionDAOImpl();
        this.deliveryDAO = new DeliveryDAOImpl();
    }

    public EventFanOutServiceImpl(SubscriptionDAO subscriptionDAO, DeliveryDAO deliveryDAO) {
        this.subscriptionDAO = subscriptionDAO;
        this.deliveryDAO = deliveryDAO;
    }

    @Override
    public void fanOutEvent(Event event, List<String> eventPurposes) {
        if (event == null)
            return;

        List<Subscription> eligibleSubscriptions = subscriptionDAO.getActiveSubscriptionsForMatching(
                event.getOrgId(), event.getGroupId(), event.getTopicId());

        Set<String> processedSubscriberGroupSet = new HashSet<>();
        Timestamp now = new Timestamp(System.currentTimeMillis());

        for (Subscription sub : eligibleSubscriptions) {
            boolean matches = FilterMatcher.isMatch(sub.getPurposeFilterMode(), sub.getPurposes(), eventPurposes);
            if (!matches) {
                continue;
            }

            // Deduplicate per subscription: each subscription gets exactly one delivery per event
            String key = sub.getSubscriptionId();
            if (processedSubscriberGroupSet.contains(key)) {
                continue;
            }
            processedSubscriberGroupSet.add(key);

            String deliveryId = UUID.randomUUID().toString();

            if ("webhook".equalsIgnoreCase(sub.getDeliveryMode())) {
                WebhookDelivery delivery = new WebhookDelivery(
                        deliveryId,
                        sub.getSubscriptionId(),
                        event.getEventId(),
                        "pending",
                        0,
                        null,
                        now,
                        now,
                        null);
                boolean saved = deliveryDAO.addWebhookDelivery(delivery);
                if (saved) {
                    LOGGER.info("Created webhook delivery record: " + deliveryId + " for subscription: "
                            + sub.getSubscriptionId());
                    // Trigger immediate webhook delivery attempt
                    WebhookDeliveryProcessor.getInstance().attemptDelivery(delivery);
                }
            } else if ("pull".equalsIgnoreCase(sub.getDeliveryMode())
                    || "poll".equalsIgnoreCase(sub.getDeliveryMode())) {
                PollDelivery pollDelivery = new PollDelivery(
                        deliveryId,
                        sub.getSubscriptionId(),
                        event.getEventId(),
                        "pending",
                        now,
                        null);
                boolean saved = deliveryDAO.addPollDelivery(pollDelivery);
                if (saved) {
                    LOGGER.info("Created poll delivery record: " + deliveryId + " for subscription: "
                            + sub.getSubscriptionId());
                }
            }
        }
    }
}
