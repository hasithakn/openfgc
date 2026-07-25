package org.wso2.dpdp.enf.endpoint.bean;

import java.util.List;

public class SubscriptionEventListResponseBean {

    private List<SubscriptionEventItemBean> content;
    private long totalElements;
    private int limit;
    private int offset;
    private int count;
    private int page;

    public SubscriptionEventListResponseBean() {
    }

    public SubscriptionEventListResponseBean(List<SubscriptionEventItemBean> content, long totalElements,
                                            int limit, int offset, int count) {
        this.content = content;
        this.totalElements = totalElements;
        this.limit = limit;
        this.offset = offset;
        this.count = count;
        this.page = limit > 0 ? offset / limit : 0;
    }

    public List<SubscriptionEventItemBean> getContent() {
        return content;
    }

    public void setContent(List<SubscriptionEventItemBean> content) {
        this.content = content;
    }

    public long getTotalElements() {
        return totalElements;
    }

    public void setTotalElements(long totalElements) {
        this.totalElements = totalElements;
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

    public int getPage() {
        return page;
    }

    public void setPage(int page) {
        this.page = page;
    }
}
