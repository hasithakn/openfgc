package org.wso2.dpdp.enf.dao.impl;

import org.wso2.dpdp.enf.dao.TopicDAO;
import org.wso2.dpdp.enf.dao.model.Topic;
import org.wso2.dpdp.enf.dao.queries.QueryConstants;
import org.wso2.dpdp.enf.dao.util.DBUtil;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.logging.Level;
import java.util.logging.Logger;

public class TopicDAOImpl implements TopicDAO {

    private static final Logger LOGGER = Logger.getLogger(TopicDAOImpl.class.getName());

    @Override
    public boolean addTopic(Topic topic) {
        Connection conn = null;
        PreparedStatement ps = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.ADD_TOPIC);
            ps.setString(1, topic.getTopicId());
            ps.setString(2, topic.getOrgId());
            ps.setString(3, topic.getName());
            ps.setString(4, topic.getDescription());
            ps.setString(5, topic.getStatus() != null ? topic.getStatus() : "active");
            return ps.executeUpdate() > 0;
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error adding topic for org: " + topic.getOrgId(), e);
        } finally {
            DBUtil.closeAll(conn, ps, null);
        }
        return false;
    }

    @Override
    public Optional<Topic> getTopicById(String topicId) {
        Connection conn = null;
        PreparedStatement ps = null;
        ResultSet rs = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.GET_TOPIC_BY_ID);
            ps.setString(1, topicId);
            rs = ps.executeQuery();
            if (rs.next()) {
                return Optional.of(mapResultSetToTopic(rs));
            }
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error getting topic by ID: " + topicId, e);
        } finally {
            DBUtil.closeAll(conn, ps, rs);
        }
        return Optional.empty();
    }

    @Override
    public Optional<Topic> getTopicByOrgAndName(String orgId, String name) {
        Connection conn = null;
        PreparedStatement ps = null;
        ResultSet rs = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.GET_TOPIC_BY_ORG_AND_NAME);
            ps.setString(1, orgId);
            ps.setString(2, name);
            rs = ps.executeQuery();
            if (rs.next()) {
                return Optional.of(mapResultSetToTopic(rs));
            }
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error getting topic by org and name: " + orgId + ", " + name, e);
        } finally {
            DBUtil.closeAll(conn, ps, rs);
        }
        return Optional.empty();
    }

    @Override
    public boolean updateTopicStatus(String topicId, String status) {
        Connection conn = null;
        PreparedStatement ps = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.UPDATE_TOPIC_STATUS);
            ps.setString(1, status);
            ps.setString(2, topicId);
            return ps.executeUpdate() > 0;
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error updating topic status: " + topicId, e);
        } finally {
            DBUtil.closeAll(conn, ps, null);
        }
        return false;
    }

    @Override
    public List<Topic> listTopics(String orgId, String status, String search, int limit, int offset, String sort,
            int[] totalOut) {
        List<Topic> topics = new ArrayList<>();
        Connection conn = null;
        PreparedStatement ps = null;
        ResultSet rs = null;

        StringBuilder sql = new StringBuilder(
                "SELECT TOPIC_ID, ORG_ID, NAME, DESCRIPTION, STATUS FROM TOPIC WHERE ORG_ID = ? ");
        List<Object> params = new ArrayList<>();
        params.add(orgId);

        if (status != null && !status.trim().isEmpty()) {
            sql.append("AND STATUS = ? ");
            params.add(status.trim());
        }
        if (search != null && !search.trim().isEmpty()) {
            sql.append("AND NAME LIKE ? ");
            params.add("%" + search.trim() + "%");
        }

        // Count Query
        String countSql = "SELECT COUNT(*) FROM (" + sql.toString() + ") AS total_count";

        // Sort
        String orderBy = "NAME ASC";
        if (sort != null && !sort.trim().isEmpty()) {
            if ("-name".equalsIgnoreCase(sort)) {
                orderBy = "NAME DESC";
            } else if ("status".equalsIgnoreCase(sort)) {
                orderBy = "STATUS ASC";
            } else if ("-status".equalsIgnoreCase(sort)) {
                orderBy = "STATUS DESC";
            }
        }
        sql.append("ORDER BY ").append(orderBy).append(" LIMIT ? OFFSET ?");

        try {
            conn = DBUtil.getConnection();

            // Execute Count
            PreparedStatement countPs = conn.prepareStatement(countSql);
            for (int i = 0; i < params.size(); i++) {
                countPs.setObject(i + 1, params.get(i));
            }
            ResultSet countRs = countPs.executeQuery();
            if (countRs.next() && totalOut != null && totalOut.length > 0) {
                totalOut[0] = countRs.getInt(1);
            }
            DBUtil.closeAll(null, countPs, countRs);

            // Execute Query
            ps = conn.prepareStatement(sql.toString());
            int idx = 1;
            for (Object param : params) {
                ps.setObject(idx++, param);
            }
            ps.setInt(idx++, limit);
            ps.setInt(idx, offset);

            rs = ps.executeQuery();
            while (rs.next()) {
                topics.add(mapResultSetToTopic(rs));
            }
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error listing topics for org: " + orgId, e);
        } finally {
            DBUtil.closeAll(conn, ps, rs);
        }
        return topics;
    }

    @Override
    public boolean hasActiveSubscriptions(String topicId) {
        Connection conn = null;
        PreparedStatement ps = null;
        ResultSet rs = null;
        String sql = "SELECT COUNT(*) FROM SUBSCRIPTION WHERE TOPIC_ID = ? AND STATUS != 'deleted'";
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(sql);
            ps.setString(1, topicId);
            rs = ps.executeQuery();
            if (rs.next()) {
                return rs.getInt(1) > 0;
            }
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error checking active subscriptions for topic: " + topicId, e);
        } finally {
            DBUtil.closeAll(conn, ps, rs);
        }
        return false;
    }

    private Topic mapResultSetToTopic(ResultSet rs) throws SQLException {
        return new Topic(
                rs.getString("TOPIC_ID"),
                rs.getString("ORG_ID"),
                rs.getString("NAME"),
                rs.getString("DESCRIPTION"),
                rs.getString("STATUS"));
    }
}
