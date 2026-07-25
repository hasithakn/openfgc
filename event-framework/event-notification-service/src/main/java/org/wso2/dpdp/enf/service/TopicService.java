package org.wso2.dpdp.enf.service;

import org.wso2.dpdp.enf.service.dto.TopicDTO;
import java.util.List;

public interface TopicService {
    TopicDTO createTopic(String orgId, String name, String description);
    List<TopicDTO> listTopics(String orgId, String status, String search, int limit, int offset, String sort, int[] totalOut);
    TopicDTO deleteTopic(String orgId, String topicIdStr);
}
