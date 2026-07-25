package org.wso2.dpdp.enf.dao.impl;

import org.wso2.dpdp.enf.dao.SubscriptionDAO;
import org.wso2.dpdp.enf.dao.model.Subscription;
import org.wso2.dpdp.enf.dao.queries.QueryConstants;
import org.wso2.dpdp.enf.dao.util.DBUtil;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Optional;
import java.util.logging.Level;
import java.util.logging.Logger;

public class SubscriptionDAOImpl implements SubscriptionDAO {

    private static final Logger LOGGER = Logger.getLogger(SubscriptionDAOImpl.class.getName());

    @Override
    public boolean addSubscription(Subscription subscription) {
        Connection conn = null;
        PreparedStatement ps = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.ADD_SUBSCRIPTION);
            ps.setString(1, subscription.getSubscriptionId());
            ps.setString(2, subscription.getOrgId());
            ps.setString(3, subscription.getGroupId());
            ps.setString(4, subscription.getTopicId());
            ps.setString(5, subscription.getStatus() != null ? subscription.getStatus() : "active");
            ps.setString(6, subscription.getPurposeFilterMode());
            ps.setString(7, subscription.getCallbackUrl());
            ps.setString(8, subscription.getSharedSecret());
            ps.setTimestamp(9, subscription.getCreatedAt() != null ? subscription.getCreatedAt() : new java.sql.Timestamp(System.currentTimeMillis()));
            ps.setTimestamp(10, subscription.getUpdatedAt() != null ? subscription.getUpdatedAt() : new java.sql.Timestamp(System.currentTimeMillis()));
            ps.setString(11, subscription.getDeliveryMode());

            int rows = ps.executeUpdate();
            if (rows > 0 && subscription.getPurposes() != null && !subscription.getPurposes().isEmpty()) {
                addSubscriptionPurposes(subscription.getSubscriptionId(), subscription.getPurposes());
            }
            return rows > 0;
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error adding subscription: " + subscription.getSubscriptionId(), e);
        } finally {
            DBUtil.closeAll(conn, ps, null);
        }
        return false;
    }

    @Override
    public Optional<Subscription> getSubscriptionById(String subscriptionId, String orgId) {
        Connection conn = null;
        PreparedStatement ps = null;
        ResultSet rs = null;
        String sql = "SELECT SUBSCRIPTION_ID, ORG_ID, GROUP_ID, TOPIC_ID, STATUS, PURPOSE_FILTER_MODE, CALLBACK_URL, " +
                "SHARED_SECRET, CREATED_AT, UPDATED_AT, DELIVERY_MODE FROM SUBSCRIPTION WHERE SUBSCRIPTION_ID = ? AND ORG_ID = ?";
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(sql);
            ps.setString(1, subscriptionId);
            ps.setString(2, orgId);
            rs = ps.executeQuery();
            if (rs.next()) {
                Subscription sub = mapResultSetToSubscription(rs);
                sub.setPurposes(getSubscriptionPurposes(subscriptionId));
                return Optional.of(sub);
            }
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error getting subscription by ID: " + subscriptionId, e);
        } finally {
            DBUtil.closeAll(conn, ps, rs);
        }
        return Optional.empty();
    }

    @Override
    public List<Subscription> getActiveSubscriptionsForMatching(String orgId, String groupId, String topicId) {
        List<Subscription> subscriptions = new ArrayList<>();
        Connection conn = null;
        PreparedStatement ps = null;
        ResultSet rs = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.GET_SUBSCRIPTIONS_FOR_MATCHING);
            ps.setString(1, orgId);
            ps.setString(2, groupId);
            ps.setString(3, topicId);
            rs = ps.executeQuery();
            while (rs.next()) {
                Subscription sub = mapResultSetToSubscription(rs);
                sub.setPurposes(getSubscriptionPurposes(sub.getSubscriptionId()));
                subscriptions.add(sub);
            }
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error getting matching subscriptions for topic: " + topicId, e);
        } finally {
            DBUtil.closeAll(conn, ps, rs);
        }
        return subscriptions;
    }

    @Override
    public boolean updateSubscriptionStatus(String subscriptionId, String status) {
        Connection conn = null;
        PreparedStatement ps = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.UPDATE_SUBSCRIPTION_STATUS);
            ps.setString(1, status);
            ps.setString(2, subscriptionId);
            return ps.executeUpdate() > 0;
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error updating subscription status: " + subscriptionId, e);
        } finally {
            DBUtil.closeAll(conn, ps, null);
        }
        return false;
    }

    @Override
    public void addSubscriptionPurposes(String subscriptionId, List<String> purposes) {
        if (purposes == null || purposes.isEmpty()) return;
        Connection conn = null;
        PreparedStatement ps = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.ADD_SUBSCRIPTION_PURPOSE);
            for (String purpose : purposes) {
                ps.setString(1, subscriptionId);
                ps.setString(2, purpose);
                ps.addBatch();
            }
            ps.executeBatch();
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error adding subscription purposes for: " + subscriptionId, e);
        } finally {
            DBUtil.closeAll(conn, ps, null);
        }
    }

    @Override
    public List<String> getSubscriptionPurposes(String subscriptionId) {
        List<String> purposes = new ArrayList<>();
        Connection conn = null;
        PreparedStatement ps = null;
        ResultSet rs = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.GET_SUBSCRIPTION_PURPOSES);
            ps.setString(1, subscriptionId);
            rs = ps.executeQuery();
            while (rs.next()) {
                purposes.add(rs.getString("PURPOSE_NAME"));
            }
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error getting subscription purposes for: " + subscriptionId, e);
        } finally {
            DBUtil.closeAll(conn, ps, rs);
        }
        return purposes;
    }

    @Override
    public List<Subscription> listSubscriptions(String orgId, String status, String purposesStr, String search, int limit, int offset, String sort, int[] totalOut) {
        List<Subscription> list = new ArrayList<>();
        Connection conn = null;
        PreparedStatement ps = null;
        ResultSet rs = null;

        StringBuilder sql = new StringBuilder("SELECT s.SUBSCRIPTION_ID, s.ORG_ID, s.GROUP_ID, s.TOPIC_ID, s.STATUS, s.PURPOSE_FILTER_MODE, ")
                .append("s.CALLBACK_URL, s.SHARED_SECRET, s.CREATED_AT, s.UPDATED_AT, s.DELIVERY_MODE FROM SUBSCRIPTION s ")
                .append("JOIN TOPIC t ON s.TOPIC_ID = t.TOPIC_ID WHERE s.ORG_ID = ? ");
        List<Object> params = new ArrayList<>();
        params.add(orgId);

        if (status != null && !status.trim().isEmpty()) {
            sql.append("AND s.STATUS = ? ");
            params.add(status.trim());
        }
        if (purposesStr != null && !purposesStr.trim().isEmpty()) {
            String[] pArray = purposesStr.split(",");
            List<String> validPurposes = new ArrayList<>();
            for (String p : pArray) {
                if (p != null && !p.trim().isEmpty()) {
                    validPurposes.add(p.trim());
                }
            }
            if (!validPurposes.isEmpty()) {
                sql.append("AND EXISTS (SELECT 1 FROM SUBSCRIPTION_PURPOSE sp WHERE sp.SUBSCRIPTION_ID = s.SUBSCRIPTION_ID AND sp.PURPOSE_NAME IN (");
                for (int i = 0; i < validPurposes.size(); i++) {
                    sql.append(i == 0 ? "?" : ", ?");
                    params.add(validPurposes.get(i));
                }
                sql.append(")) ");
            }
        }
        if (search != null && !search.trim().isEmpty()) {
            sql.append("AND (s.GROUP_ID LIKE ? OR s.STATUS LIKE ? OR t.NAME LIKE ? OR EXISTS (SELECT 1 FROM SUBSCRIPTION_PURPOSE sp WHERE sp.SUBSCRIPTION_ID = s.SUBSCRIPTION_ID AND sp.PURPOSE_NAME LIKE ?)) ");
            String term = "%" + search.trim() + "%";
            params.add(term);
            params.add(term);
            params.add(term);
            params.add(term);
        }

        String countSql = "SELECT COUNT(DISTINCT SUBSCRIPTION_ID) FROM (" + sql.toString() + ") AS total_sub";

        String orderBy = "s.UPDATED_AT DESC";
        if (sort != null && !sort.trim().isEmpty()) {
            if ("createdAt".equalsIgnoreCase(sort)) orderBy = "s.CREATED_AT ASC";
            else if ("-createdAt".equalsIgnoreCase(sort)) orderBy = "s.CREATED_AT DESC";
            else if ("updatedAt".equalsIgnoreCase(sort)) orderBy = "s.UPDATED_AT ASC";
            else if ("-updatedAt".equalsIgnoreCase(sort)) orderBy = "s.UPDATED_AT DESC";
        }
        sql.append("ORDER BY ").append(orderBy).append(" LIMIT ? OFFSET ?");

        try {
            conn = DBUtil.getConnection();

            PreparedStatement countPs = conn.prepareStatement(countSql);
            for (int i = 0; i < params.size(); i++) {
                countPs.setObject(i + 1, params.get(i));
            }
            ResultSet countRs = countPs.executeQuery();
            if (countRs.next() && totalOut != null && totalOut.length > 0) {
                totalOut[0] = countRs.getInt(1);
            }
            DBUtil.closeAll(null, countPs, countRs);

            ps = conn.prepareStatement(sql.toString());
            int idx = 1;
            for (Object param : params) {
                ps.setObject(idx++, param);
            }
            ps.setInt(idx++, limit);
            ps.setInt(idx, offset);

            rs = ps.executeQuery();
            while (rs.next()) {
                Subscription sub = mapResultSetToSubscription(rs);
                sub.setPurposes(getSubscriptionPurposes(sub.getSubscriptionId()));
                list.add(sub);
            }
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error listing subscriptions for org: " + orgId, e);
        } finally {
            DBUtil.closeAll(conn, ps, rs);
        }
        return list;
    }

    @Override
    public boolean hasPendingDeliveries(String subscriptionId) {
        Connection conn = null;
        PreparedStatement ps = null;
        ResultSet rs = null;
        String sql = "SELECT (SELECT COUNT(*) FROM WEBHOOK_DELIVERY WHERE SUBSCRIPTION_ID = ? AND STATUS IN ('pending', 'delivered')) + " +
                "(SELECT COUNT(*) FROM POLL_DELIVERY WHERE SUBSCRIPTION_ID = ? AND STATUS = 'pending') AS total_pending";
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(sql);
            ps.setString(1, subscriptionId);
            ps.setString(2, subscriptionId);
            rs = ps.executeQuery();
            if (rs.next()) {
                return rs.getInt(1) > 0;
            }
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error checking pending deliveries for subscription: " + subscriptionId, e);
        } finally {
            DBUtil.closeAll(conn, ps, rs);
        }
        return false;
    }

    @Override
    public Optional<Subscription> findDuplicateSubscription(String orgId, String groupId, String topicId, String purposeFilterMode, List<String> sortedPurposes) {
        List<Subscription> activeSubs = getActiveSubscriptionsForMatching(orgId, groupId, topicId);
        for (Subscription sub : activeSubs) {
            if (sub.getPurposeFilterMode().equalsIgnoreCase(purposeFilterMode)) {
                List<String> existingPurposes = sub.getPurposes() != null ? new ArrayList<>(sub.getPurposes()) : Collections.emptyList();
                Collections.sort(existingPurposes);
                List<String> targetPurposes = sortedPurposes != null ? new ArrayList<>(sortedPurposes) : Collections.emptyList();
                Collections.sort(targetPurposes);
                if (existingPurposes.equals(targetPurposes)) {
                    return Optional.of(sub);
                }
            }
        }
        return Optional.empty();
    }

    @Override
    public Optional<Subscription> getActivePullSubscription(String orgId, String groupId) {
        Connection conn = null;
        PreparedStatement ps = null;
        ResultSet rs = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.GET_ACTIVE_PULL_SUBSCRIPTION_BY_GROUP);
            ps.setString(1, orgId);
            ps.setString(2, groupId);
            rs = ps.executeQuery();
            if (rs.next()) {
                Subscription sub = mapResultSetToSubscription(rs);
                sub.setPurposes(getSubscriptionPurposes(sub.getSubscriptionId()));
                return Optional.of(sub);
            }
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error getting active pull subscription for org: " + orgId + ", group: " + groupId, e);
        } finally {
            DBUtil.closeAll(conn, ps, rs);
        }
        return Optional.empty();
    }

    private Subscription mapResultSetToSubscription(ResultSet rs) throws SQLException {
        return new Subscription(
                rs.getString("SUBSCRIPTION_ID"),
                rs.getString("ORG_ID"),
                rs.getString("GROUP_ID"),
                rs.getString("TOPIC_ID"),
                rs.getString("STATUS"),
                rs.getString("PURPOSE_FILTER_MODE"),
                rs.getString("CALLBACK_URL"),
                rs.getString("SHARED_SECRET"),
                rs.getTimestamp("CREATED_AT"),
                rs.getTimestamp("UPDATED_AT"),
                rs.getString("DELIVERY_MODE")
        );
    }
}
