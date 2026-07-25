package org.wso2.dpdp.enf.dao.model;

public class Topic {

    private String topicId;
    private String orgId;
    private String name;
    private String description;
    private String status;

    public Topic() {
    }

    public Topic(String topicId, String orgId, String name, String description, String status) {
        this.topicId = topicId;
        this.orgId = orgId;
        this.name = name;
        this.description = description;
        this.status = status;
    }

    public String getTopicId() {
        return topicId;
    }

    public void setTopicId(String topicId) {
        this.topicId = topicId;
    }

    public String getOrgId() {
        return orgId;
    }

    public void setOrgId(String orgId) {
        this.orgId = orgId;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }
}
