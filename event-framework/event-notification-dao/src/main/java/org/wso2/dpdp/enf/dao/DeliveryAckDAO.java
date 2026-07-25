package org.wso2.dpdp.enf.dao;

import org.wso2.dpdp.enf.dao.model.WebhookDeliveryAck;
import java.util.Optional;

public interface DeliveryAckDAO {
    boolean addDeliveryAck(WebhookDeliveryAck ack);
    Optional<WebhookDeliveryAck> getDeliveryAckByDeliveryId(String deliveryId);
}
