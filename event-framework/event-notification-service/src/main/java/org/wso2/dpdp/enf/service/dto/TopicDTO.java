package org.wso2.dpdp.enf.service.dto;

public class TopicDTO {

    private String topicId;
    private String name;
    private String description;
    private String status;

    public TopicDTO() {
    }

    public TopicDTO(String topicId, String name, String description, String status) {
        this.topicId = topicId;
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
