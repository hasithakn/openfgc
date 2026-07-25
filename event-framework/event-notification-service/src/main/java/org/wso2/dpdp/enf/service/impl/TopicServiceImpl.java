package org.wso2.dpdp.enf.service.impl;

import org.wso2.dpdp.enf.dao.TopicDAO;
import org.wso2.dpdp.enf.dao.impl.TopicDAOImpl;
import org.wso2.dpdp.enf.dao.model.Topic;
import org.wso2.dpdp.enf.service.TopicService;
import org.wso2.dpdp.enf.service.dto.TopicDTO;
import org.wso2.dpdp.enf.service.exception.ENFException;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

public class TopicServiceImpl implements TopicService {

    private final TopicDAO topicDAO;

    public TopicServiceImpl() {
        this.topicDAO = new TopicDAOImpl();
    }

    public TopicServiceImpl(TopicDAO topicDAO) {
        this.topicDAO = topicDAO;
    }

    @Override
    public TopicDTO createTopic(String orgId, String name, String description) {
        if (name == null || name.trim().isEmpty()) {
            throw new ENFException("CS-4001", "Malformed request", "Topic name is required.", 400);
        }

        // Check if topic name already exists for this org
        Optional<Topic> existing = topicDAO.getTopicByOrgAndName(orgId, name.trim());
        if (existing.isPresent()) {
            throw new ENFException("CS-4090", "Topic already exists", "A topic with this name already exists for this org.", 409);
        }

        String topicId = UUID.randomUUID().toString();
        Topic topic = new Topic(topicId, orgId, name.trim(), description != null ? description.trim() : null, "active");
        boolean created = topicDAO.addTopic(topic);
        if (!created) {
            throw new ENFException("CS-5000", "Internal error", "Failed to create topic.", 500);
        }

        return new TopicDTO(topicId, topic.getName(), topic.getDescription(), "active");
    }

    @Override
    public List<TopicDTO> listTopics(String orgId, String status, String search, int limit, int offset, String sort, int[] totalOut) {
        List<Topic> topics = topicDAO.listTopics(orgId, status, search, limit, offset, sort, totalOut);
        List<TopicDTO> dtoList = new ArrayList<>();
        for (Topic t : topics) {
            dtoList.add(new TopicDTO(t.getTopicId(), t.getName(), t.getDescription(), t.getStatus()));
        }
        return dtoList;
    }

    @Override
    public TopicDTO deleteTopic(String orgId, String topicIdStr) {
        if (topicIdStr == null || topicIdStr.trim().isEmpty()) {
            throw new ENFException("CS-4040", "Resource not found", "Topic ID not found.", 404);
        }
        String topicId = topicIdStr.trim();

        Optional<Topic> topicOpt = topicDAO.getTopicById(topicId);
        if (topicOpt.isEmpty()
                || orgId == null
                || !topicOpt.get().getOrgId().equalsIgnoreCase(orgId.trim())
                || "deregistered".equalsIgnoreCase(topicOpt.get().getStatus())) {
            throw new ENFException("CS-4040", "Resource not found", "No topic exists with this ID for the given org.", 404);
        }

        // Check if active subscriptions reference it
        if (topicDAO.hasActiveSubscriptions(topicId)) {
            throw new ENFException("CS-4091", "Topic has active subscriptions", "This topic cannot be deregistered while active subscriptions reference it.", 409);
        }

        boolean updated = topicDAO.updateTopicStatus(topicId, "deregistered");
        if (!updated) {
            throw new ENFException("CS-5000", "Internal error", "Failed to update topic status.", 500);
        }

        Topic topic = topicOpt.get();
        return new TopicDTO(topic.getTopicId(), topic.getName(), topic.getDescription(), "deregistered");
    }
}
