package org.wso2.dpdp.enf.endpoint.bean;

import java.util.List;

public class TopicListResponseBean {

    private List<TopicResponseBean> topics;
    private int total;
    private int limit;
    private int offset;
    private int count;

    public TopicListResponseBean() {
    }

    public TopicListResponseBean(List<TopicResponseBean> topics, int total, int limit, int offset, int count) {
        this.topics = topics;
        this.total = total;
        this.limit = limit;
        this.offset = offset;
        this.count = count;
    }

    public List<TopicResponseBean> getTopics() {
        return topics;
    }

    public void setTopics(List<TopicResponseBean> topics) {
        this.topics = topics;
    }

    public int getTotal() {
        return total;
    }

    public void setTotal(int total) {
        this.total = total;
    }

    public int getLimit() {
        return limit;
    }

    public void setLimit(int limit) {
        this.limit = limit;
    }

    public int getOffset() {
        return offset;
    }

    public void setOffset(int offset) {
        this.offset = offset;
    }

    public int getCount() {
        return count;
    }

    public void setCount(int count) {
        this.count = count;
    }
}
