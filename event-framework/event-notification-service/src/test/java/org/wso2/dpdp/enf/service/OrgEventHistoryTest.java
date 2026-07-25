package org.wso2.dpdp.enf.service;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.wso2.dpdp.enf.dao.DeliveryAckDAO;
import org.wso2.dpdp.enf.dao.DeliveryDAO;
import org.wso2.dpdp.enf.dao.EventDAO;
import org.wso2.dpdp.enf.dao.TopicDAO;
import org.wso2.dpdp.enf.dao.model.PollDelivery;
import org.wso2.dpdp.enf.dao.model.SubscriptionDeliverySummary;
import org.wso2.dpdp.enf.dao.model.WebhookDelivery;
import org.wso2.dpdp.enf.dao.model.WebhookDeliveryAck;
import org.wso2.dpdp.enf.dao.model.WebhookDeliveryAudit;
import org.wso2.dpdp.enf.service.dto.SubscriptionDeliveryDTO;
import org.wso2.dpdp.enf.service.dto.SubscriptionEventHistoryDTO;
import org.wso2.dpdp.enf.service.exception.ENFException;
import org.wso2.dpdp.enf.service.impl.EventPublishServiceImpl;

import java.sql.Timestamp;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

public class OrgEventHistoryTest {

    private EventDAO eventDAO;
    private TopicDAO topicDAO;
    private EventFanOutService fanOutService;
    private DeliveryDAO deliveryDAO;
    private DeliveryAckDAO deliveryAckDAO;
    private EventPublishService service;

    private final String orgId = "org-001";

    @BeforeEach
    public void setUp() {
        eventDAO = mock(EventDAO.class);
        topicDAO = mock(TopicDAO.class);
        fanOutService = mock(EventFanOutService.class);
        deliveryDAO = mock(DeliveryDAO.class);
        deliveryAckDAO = mock(DeliveryAckDAO.class);

        service = new EventPublishServiceImpl(eventDAO, topicDAO, fanOutService, deliveryDAO, deliveryAckDAO);
    }

    @Test
    public void testListOrgEvents_StatusFilterAndSearch() {
        List<SubscriptionDeliverySummary> mockSummaries = new ArrayList<>();
        mockSummaries.add(new SubscriptionDeliverySummary("del-wh-1", "evt-101", "sub-123", "consent.revoke", "delivered", "WEBHOOK", new Timestamp(1700000000000L), new Timestamp(1700000000000L)));

        when(deliveryDAO.listOrgDeliveries(eq(orgId), eq("delivered"), eq("sub-123"), eq("marketing"), eq("consent"), eq(20), eq(0), any(int[].class)))
                .thenAnswer(invocation -> {
                    int[] totalOut = invocation.getArgument(7);
                    if (totalOut != null && totalOut.length > 0) {
                        totalOut[0] = 1;
                    }
                    return mockSummaries;
                });

        int[] totalOut = new int[]{0};
        List<SubscriptionDeliveryDTO> result = service.listOrgEvents(orgId, "delivered", "sub-123", "marketing", "consent", 20, 0, totalOut);

        assertEquals(1, result.size());
        assertEquals("del-wh-1", result.get(0).getDeliveryId());
        assertEquals("consent.revoke", result.get(0).getTopic());
        assertEquals("DELIVERED", result.get(0).getCurrentStatus());
        assertEquals("WEBHOOK", result.get(0).getDeliveryMode());
        assertEquals(1, totalOut[0]);
    }

    @Test
    public void testListOrgEvents_MissingOrgId_Throws400() {
        ENFException ex = assertThrows(ENFException.class, () -> {
            service.listOrgEvents("", null, null, null, null, 20, 0, new int[]{0});
        });

        assertEquals(400, ex.getStatusCode());
        assertEquals("CS-4001", ex.getCode());
    }

    @Test
    public void testGetOrgEventHistory_WebhookDetails() {
        String deliveryId = "del-wh-101";
        SubscriptionDeliverySummary summary = new SubscriptionDeliverySummary(deliveryId, "evt-101", "data.erasure", "delivered", "WEBHOOK", new Timestamp(1700000000000L), new Timestamp(1700000000000L));
        when(deliveryDAO.getOrgDeliveryById(eq(orgId), eq(deliveryId))).thenReturn(Optional.of(summary));

        WebhookDelivery wh = new WebhookDelivery(deliveryId, "sub-123", "evt-101", "delivered", 1, null, new Timestamp(1700000000000L), new Timestamp(1700000010000L), new Timestamp(1700000010000L));
        when(deliveryDAO.getWebhookDeliveryById(eq(deliveryId), eq(orgId))).thenReturn(Optional.of(wh));

        List<WebhookDeliveryAudit> audits = new ArrayList<>();
        audits.add(new WebhookDeliveryAudit("aud-1", "evt-101", deliveryId, orgId, "200", new Timestamp(1700000010000L), new Timestamp(1700000010000L)));
        when(deliveryDAO.getWebhookDeliveryAudits(eq(deliveryId))).thenReturn(audits);

        WebhookDeliveryAck ack = new WebhookDeliveryAck("ack-1", deliveryId, new Timestamp(1700000011000L), "completed", "PROOF-888");
        when(deliveryAckDAO.getDeliveryAckByDeliveryId(eq(deliveryId))).thenReturn(Optional.of(ack));

        SubscriptionEventHistoryDTO dto = service.getOrgEventHistory(orgId, deliveryId);

        assertNotNull(dto);
        assertEquals(deliveryId, dto.getDeliveryId());
        assertEquals("data.erasure", dto.getTopic());
        assertEquals("WEBHOOK", dto.getDeliveryMode());
        assertEquals("DELIVERED", dto.getCurrentStatus());
        assertEquals("COMPLETED", dto.getCompletionStatus());
        assertEquals("PROOF-888", dto.getCompletionEvidence());
        assertEquals(1, dto.getHistory().size());
        assertEquals(200, dto.getHistory().get(0).getHttpStatus());
    }

    @Test
    public void testGetOrgEventHistory_UnownedDelivery_Throws404() {
        when(deliveryDAO.getOrgDeliveryById(eq(orgId), eq("invalid-del"))).thenReturn(Optional.empty());

        ENFException ex = assertThrows(ENFException.class, () -> {
            service.getOrgEventHistory(orgId, "invalid-del");
        });

        assertEquals(404, ex.getStatusCode());
        assertEquals("CS-4042", ex.getCode());
    }
}
