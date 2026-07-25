package org.wso2.dpdp.enf.dao;

import org.wso2.dpdp.enf.dao.model.Topic;
import java.util.List;
import java.util.Optional;

public interface TopicDAO {
    boolean addTopic(Topic topic);
    Optional<Topic> getTopicById(String topicId);
    Optional<Topic> getTopicByOrgAndName(String orgId, String name);
    boolean updateTopicStatus(String topicId, String status);
    List<Topic> listTopics(String orgId, String status, String search, int limit, int offset, String sort, int[] totalOut);
    boolean hasActiveSubscriptions(String topicId);
}
