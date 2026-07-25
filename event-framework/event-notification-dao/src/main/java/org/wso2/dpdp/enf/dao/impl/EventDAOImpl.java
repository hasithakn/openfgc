package org.wso2.dpdp.enf.dao.impl;

import org.wso2.dpdp.enf.dao.EventDAO;
import org.wso2.dpdp.enf.dao.model.Event;
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

public class EventDAOImpl implements EventDAO {

    private static final Logger LOGGER = Logger.getLogger(EventDAOImpl.class.getName());

    @Override
    public boolean addEvent(Event event) {
        Connection conn = null;
        PreparedStatement ps = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.ADD_EVENT);
            ps.setString(1, event.getEventId());
            ps.setString(2, event.getOrgId());
            ps.setString(3, event.getGroupId());
            ps.setString(4, event.getTopicId());
            ps.setString(5, event.getPayload());
            ps.setTimestamp(6, event.getCreatedAt() != null ? event.getCreatedAt() : new java.sql.Timestamp(System.currentTimeMillis()));

            int rows = ps.executeUpdate();
            if (rows > 0 && event.getPurposes() != null && !event.getPurposes().isEmpty()) {
                addEventPurposes(event.getEventId(), event.getPurposes());
            }
            return rows > 0;
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error adding event: " + event.getEventId(), e);
        } finally {
            DBUtil.closeAll(conn, ps, null);
        }
        return false;
    }

    @Override
    public Optional<Event> getEventById(String eventId) {
        Connection conn = null;
        PreparedStatement ps = null;
        ResultSet rs = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.GET_EVENT_BY_ID);
            ps.setString(1, eventId);
            rs = ps.executeQuery();
            if (rs.next()) {
                Event event = new Event(
                        rs.getString("EVENT_ID"),
                        rs.getString("ORG_ID"),
                        rs.getString("GROUP_ID"),
                        rs.getString("TOPIC_ID"),
                        rs.getString("PAYLOAD"),
                        rs.getTimestamp("CREATED_AT")
                );
                event.setPurposes(getEventPurposes(eventId));
                return Optional.of(event);
            }
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error getting event by ID: " + eventId, e);
        } finally {
            DBUtil.closeAll(conn, ps, rs);
        }
        return Optional.empty();
    }

    @Override
    public void addEventPurposes(String eventId, List<String> purposes) {
        if (purposes == null || purposes.isEmpty()) return;
        Connection conn = null;
        PreparedStatement ps = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.ADD_EVENT_PURPOSE);
            for (String purpose : purposes) {
                ps.setString(1, eventId);
                ps.setString(2, purpose);
                ps.addBatch();
            }
            ps.executeBatch();
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error adding event purposes for: " + eventId, e);
        } finally {
            DBUtil.closeAll(conn, ps, null);
        }
    }

    @Override
    public List<String> getEventPurposes(String eventId) {
        List<String> purposes = new ArrayList<>();
        Connection conn = null;
        PreparedStatement ps = null;
        ResultSet rs = null;
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(QueryConstants.GET_EVENT_PURPOSES);
            ps.setString(1, eventId);
            rs = ps.executeQuery();
            while (rs.next()) {
                purposes.add(rs.getString("PURPOSE_NAME"));
            }
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error getting event purposes for: " + eventId, e);
        } finally {
            DBUtil.closeAll(conn, ps, rs);
        }
        return purposes;
    }

    @Override
    public boolean hasActiveEventsForTopic(String topicId) {
        Connection conn = null;
        PreparedStatement ps = null;
        ResultSet rs = null;
        String sql = "SELECT COUNT(*) FROM EVENT WHERE TOPIC_ID = ?";
        try {
            conn = DBUtil.getConnection();
            ps = conn.prepareStatement(sql);
            ps.setString(1, topicId);
            rs = ps.executeQuery();
            if (rs.next()) {
                return rs.getInt(1) > 0;
            }
        } catch (SQLException e) {
            LOGGER.log(Level.SEVERE, "Error checking active events for topic: " + topicId, e);
        } finally {
            DBUtil.closeAll(conn, ps, rs);
        }
        return false;
    }
}
