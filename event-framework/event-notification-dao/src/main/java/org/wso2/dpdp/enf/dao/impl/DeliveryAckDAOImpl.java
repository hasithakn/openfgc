package org.wso2.dpdp.enf.dao.impl;

import org.wso2.dpdp.enf.dao.DeliveryAckDAO;
import org.wso2.dpdp.enf.dao.model.WebhookDeliveryAck;
import org.wso2.dpdp.enf.dao.queries.QueryConstants;
import org.wso2.dpdp.enf.dao.util.DBUtil;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.Optional;
import java.util.logging.Level;
import java.util.logging.Logger;

public class DeliveryAckDAOImpl implements DeliveryAckDAO {

    private static final Logger LOGGER = Logger.getLogger(DeliveryAckDAOImpl.class.getName());

    @Override
    public boolean addDeliveryAck(WebhookDeliveryAck ack) {
        Connection conn = null;
        PreparedStatement ps = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.ADD_WEBHOOK_DELIVERY_ACK);
            ps.setString(1, ack.getAckId());
            ps.setString(2, ack.getDeliveryId());
            ps.setTimestamp(3, ack.getCompletedAt() != null ? ack.getCompletedAt() : new java.sql.Timestamp(System.currentTimeMillis()));
            ps.setString(4, ack.getCompletionStatus());
            ps.setString(5, ack.getCompletionEvidence());
            return ps.executeUpdate() > 0;
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error adding delivery ACK for delivery: " + ack.getDeliveryId(), e);
        } finally {
            DBUtil.closeAll(conn, ps, null);
        }
        return false;
    }

    @Override
    public Optional<WebhookDeliveryAck> getDeliveryAckByDeliveryId(String deliveryId) {
        Connection conn = null;
        PreparedStatement ps = null;
        ResultSet rs = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.GET_WEBHOOK_DELIVERY_ACK_BY_DELIVERY_ID);
            ps.setString(1, deliveryId);
            rs = ps.executeQuery();
            if (rs.next()) {
                return Optional.of(new WebhookDeliveryAck(
                        rs.getString("ACK_ID"),
                        rs.getString("DELIVERY_ID"),
                        rs.getTimestamp("COMPLETED_AT"),
                        rs.getString("COMPLETION_STATUS"),
                        rs.getString("COMPLETION_EVIDENCE")
                ));
            }
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error getting delivery ACK for delivery: " + deliveryId, e);
        } finally {
            DBUtil.closeAll(conn, ps, rs);
        }
        return Optional.empty();
    }
}
