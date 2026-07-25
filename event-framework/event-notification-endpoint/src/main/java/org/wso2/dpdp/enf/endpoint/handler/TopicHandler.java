package org.wso2.dpdp.enf.endpoint.handler;

import org.wso2.dpdp.enf.endpoint.bean.TopicCreateRequestBean;
import org.wso2.dpdp.enf.endpoint.bean.TopicListResponseBean;
import org.wso2.dpdp.enf.endpoint.bean.TopicResponseBean;
import org.wso2.dpdp.enf.service.TopicService;
import org.wso2.dpdp.enf.service.dto.TopicDTO;
import org.wso2.dpdp.enf.service.impl.TopicServiceImpl;

import java.util.ArrayList;
import java.util.List;

public class TopicHandler {

    private final TopicService topicService;

    public TopicHandler() {
        this.topicService = new TopicServiceImpl();
    }

    public TopicHandler(TopicService topicService) {
        this.topicService = topicService;
    }

    public TopicResponseBean createTopic(String orgId, TopicCreateRequestBean request) {
        String name = request != null ? request.getName() : null;
        String desc = request != null ? request.getDescription() : null;
        TopicDTO dto = topicService.createTopic(orgId, name, desc);
        return new TopicResponseBean(dto.getTopicId(), dto.getName(), dto.getDescription(), dto.getStatus());
    }

    public TopicListResponseBean listTopics(String orgId, String status, String search, Integer limit, Integer offset, Integer count, String sort) {
        int lim = limit != null && limit > 0 ? limit : 20;
        int off = offset != null && offset >= 0 ? offset : 0;
        int[] totalOut = new int[]{0};

        List<TopicDTO> list = topicService.listTopics(orgId, status, search, lim, off, sort, totalOut);
        List<TopicResponseBean> beanList = new ArrayList<>();
        for (TopicDTO dto : list) {
            beanList.add(new TopicResponseBean(dto.getTopicId(), dto.getName(), dto.getDescription(), dto.getStatus()));
        }
        return new TopicListResponseBean(beanList, totalOut[0], lim, off, beanList.size());
    }

    public TopicResponseBean deleteTopic(String orgId, String topicId) {
        TopicDTO dto = topicService.deleteTopic(orgId, topicId);
        return new TopicResponseBean(dto.getTopicId(), dto.getName(), dto.getDescription(), dto.getStatus());
    }
}
