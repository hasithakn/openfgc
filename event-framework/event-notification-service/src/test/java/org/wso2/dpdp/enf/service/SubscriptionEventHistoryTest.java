package org.wso2.dpdp.enf.service;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.wso2.dpdp.enf.dao.DeliveryAckDAO;
import org.wso2.dpdp.enf.dao.DeliveryDAO;
import org.wso2.dpdp.enf.dao.SubscriptionDAO;
import org.wso2.dpdp.enf.dao.TopicDAO;
import org.wso2.dpdp.enf.dao.model.PollDelivery;
import org.wso2.dpdp.enf.dao.model.Subscription;
import org.wso2.dpdp.enf.dao.model.SubscriptionDeliverySummary;
import org.wso2.dpdp.enf.dao.model.WebhookDelivery;
import org.wso2.dpdp.enf.dao.model.WebhookDeliveryAck;
import org.wso2.dpdp.enf.dao.model.WebhookDeliveryAudit;
import org.wso2.dpdp.enf.service.dto.SubscriptionDeliveryDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionEventHistoryDTO;
import org.wso2.dpdp.enf.service.exception.ENFException;
import org.wso2.dpdp.enf.service.impl.SubscriptionServiceImpl;

import java.sql.Timestamp;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

public class SubscriptionEventHistoryTest {

    private SubscriptionDAO subscriptionDAO;
    private TopicDAO topicDAO;
    private DeliveryDAO deliveryDAO;
    private DeliveryAckDAO deliveryAckDAO;
    private SubscriptionService service;

    private final String orgId = "org-001";
    private final String subId = "sub-123";

    @BeforeEach
    public void setUp() {
        subscriptionDAO = mock(SubscriptionDAO.class);
        topicDAO = mock(TopicDAO.class);
        deliveryDAO = mock(DeliveryDAO.class);
        deliveryAckDAO = mock(DeliveryAckDAO.class);

        service = new SubscriptionServiceImpl(subscriptionDAO, topicDAO, deliveryDAO, deliveryAckDAO);

        Subscription sub = new Subscription(subId, orgId, "group-1", "topic-1", "active", "all", "https://example.com/wh", "sec", new Timestamp(System.currentTimeMillis()), new Timestamp(System.currentTimeMillis()), "webhook");
        when(subscriptionDAO.getSubscriptionById(eq(subId), eq(orgId))).thenReturn(Optional.of(sub));
    }

    @Test
    public void testListSubscriptionEvents_WebhookAndPollCombined() {
        List<SubscriptionDeliverySummary> mockSummaries = new ArrayList<>();
        mockSummaries.add(new SubscriptionDeliverySummary("del-wh-1", "evt-101", "consent.revoke", "delivered", "WEBHOOK", new Timestamp(1700000000000L), new Timestamp(1700000000000L)));
        mockSummaries.add(new SubscriptionDeliverySummary("del-poll-1", "evt-102", "consent.granted", "pending", "POLL", new Timestamp(1700000050000L), new Timestamp(1700000050000L)));

        when(deliveryDAO.listSubscriptionDeliveries(eq(subId), eq(20), eq(0), any(int[].class)))
                .thenAnswer(invocation -> {
                    int[] totalOut = invocation.getArgument(3);
                    if (totalOut != null && totalOut.length > 0) {
                        totalOut[0] = 2;
                    }
                    return mockSummaries;
                });

        int[] totalOut = new int[]{0};
        List<SubscriptionDeliveryDTO> result = service.listSubscriptionEvents(orgId, subId, 20, 0, totalOut);

        assertEquals(2, result.size());
        assertEquals("del-wh-1", result.get(0).getDeliveryId());
        assertEquals("consent.revoke", result.get(0).getTopic());
        assertEquals("WEBHOOK", result.get(0).getDeliveryMode());

        assertEquals("del-poll-1", result.get(1).getDeliveryId());
        assertEquals("consent.granted", result.get(1).getTopic());
        assertEquals("POLL", result.get(1).getDeliveryMode());
        assertEquals(2, totalOut[0]);
    }

    @Test
    public void testListSubscriptionEvents_EmptyHistory() {
        when(deliveryDAO.listSubscriptionDeliveries(eq(subId), anyInt(), anyInt(), any(int[].class)))
                .thenReturn(new ArrayList<>());

        int[] totalOut = new int[]{0};
        List<SubscriptionDeliveryDTO> result = service.listSubscriptionEvents(orgId, subId, 20, 0, totalOut);

        assertNotNull(result);
        assertTrue(result.isEmpty());
        assertEquals(0, totalOut[0]);
    }

    @Test
    public void testListSubscriptionEvents_MissingSubscription_Throws404() {
        when(subscriptionDAO.getSubscriptionById(eq("unknown-sub"), eq(orgId))).thenReturn(Optional.empty());

        ENFException ex = assertThrows(ENFException.class, () -> {
            service.listSubscriptionEvents(orgId, "unknown-sub", 20, 0, new int[]{0});
        });

        assertEquals(404, ex.getStatusCode());
        assertEquals("CS-4040", ex.getCode());
    }

    @Test
    public void testGetSubscriptionEventHistory_WebhookRetryAndAck() {
        String deliveryId = "del-wh-1";
        SubscriptionDeliverySummary summary = new SubscriptionDeliverySummary(deliveryId, "evt-101", "user.created", "delivered", "WEBHOOK", new Timestamp(1700000000000L), new Timestamp(1700000000000L));
        when(deliveryDAO.getSubscriptionDeliveryById(eq(subId), eq(deliveryId))).thenReturn(Optional.of(summary));

        WebhookDelivery wh = new WebhookDelivery(deliveryId, subId, "evt-101", "delivered", 2, null, new Timestamp(1700000000000L), new Timestamp(1700000060000L), new Timestamp(1700000060000L));
        when(deliveryDAO.getWebhookDeliveryById(eq(deliveryId), eq(orgId))).thenReturn(Optional.of(wh));

        List<WebhookDeliveryAudit> audits = new ArrayList<>();
        audits.add(new WebhookDeliveryAudit("aud-1", "evt-101", deliveryId, orgId, "500", new Timestamp(1700000000000L), new Timestamp(1700000000000L)));
        audits.add(new WebhookDeliveryAudit("aud-2", "evt-101", deliveryId, orgId, "200", new Timestamp(1700000060000L), new Timestamp(1700000060000L)));
        when(deliveryDAO.getWebhookDeliveryAudits(eq(deliveryId))).thenReturn(audits);

        WebhookDeliveryAck ack = new WebhookDeliveryAck("ack-1", deliveryId, new Timestamp(1700000061000L), "completed", "JWT-PROOF-123");
        when(deliveryAckDAO.getDeliveryAckByDeliveryId(eq(deliveryId))).thenReturn(Optional.of(ack));

        SubscriptionEventHistoryDTO dto = service.getSubscriptionEventHistory(orgId, subId, deliveryId);

        assertNotNull(dto);
        assertEquals(deliveryId, dto.getDeliveryId());
        assertEquals("evt-101", dto.getEventId());
        assertEquals("user.created", dto.getTopic());
        assertEquals("WEBHOOK", dto.getDeliveryMode());
        assertEquals("DELIVERED", dto.getCurrentStatus());
        assertEquals("COMPLETED", dto.getCompletionStatus());
        assertEquals("JWT-PROOF-123", dto.getCompletionEvidence());

        assertEquals(2, dto.getHistory().size());
        assertEquals(1, dto.getHistory().get(0).getAttempt());
        assertEquals("FAILED", dto.getHistory().get(0).getStatus());
        assertEquals(500, dto.getHistory().get(0).getHttpStatus());

        assertEquals(2, dto.getHistory().get(1).getAttempt());
        assertEquals("DELIVERED", dto.getHistory().get(1).getStatus());
        assertEquals(200, dto.getHistory().get(1).getHttpStatus());
    }

    @Test
    public void testGetSubscriptionEventHistory_PollCompletion() {
        String deliveryId = "del-poll-1";
        SubscriptionDeliverySummary summary = new SubscriptionDeliverySummary(deliveryId, "evt-202", "order.placed", "acknowledged", "POLL", new Timestamp(1700000000000L), new Timestamp(1700000000000L));
        when(deliveryDAO.getSubscriptionDeliveryById(eq(subId), eq(deliveryId))).thenReturn(Optional.of(summary));

        PollDelivery pd = new PollDelivery(deliveryId, subId, "evt-202", "acknowledged", new Timestamp(1700000000000L), new Timestamp(1700000090000L));
        when(deliveryDAO.getPollDeliveryById(eq(deliveryId), eq(orgId))).thenReturn(Optional.of(pd));

        SubscriptionEventHistoryDTO dto = service.getSubscriptionEventHistory(orgId, subId, deliveryId);

        assertNotNull(dto);
        assertEquals(deliveryId, dto.getDeliveryId());
        assertEquals("POLL", dto.getDeliveryMode());
        assertEquals("ACKNOWLEDGED", dto.getCurrentStatus());
        assertEquals("ACKNOWLEDGED", dto.getCompletionStatus());
        assertEquals(1, dto.getHistory().size());
        assertEquals(1700000090000L, dto.getHistory().get(0).getTimestamp());
    }

    @Test
    public void testGetSubscriptionEventHistory_UnownedDelivery_Throws404() {
        when(deliveryDAO.getSubscriptionDeliveryById(eq(subId), eq("other-delivery"))).thenReturn(Optional.empty());

        ENFException ex = assertThrows(ENFException.class, () -> {
            service.getSubscriptionEventHistory(orgId, subId, "other-delivery");
        });

        assertEquals(404, ex.getStatusCode());
        assertEquals("CS-4042", ex.getCode());
    }
}
