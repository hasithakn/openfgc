package org.wso2.dpdp.enf.endpoint.bean;

public class CompletionRequestBean {

    private String completionStatus;
    private String completionEvidence;
    private Long completedAt;

    public CompletionRequestBean() {
    }

    public CompletionRequestBean(String completionStatus, String completionEvidence, Long completedAt) {
        this.completionStatus = completionStatus;
        this.completionEvidence = completionEvidence;
        this.completedAt = completedAt;
    }

    public String getCompletionStatus() {
        return completionStatus;
    }

    public void setCompletionStatus(String completionStatus) {
        this.completionStatus = completionStatus;
    }

    public String getCompletionEvidence() {
        return completionEvidence;
    }

    public void setCompletionEvidence(String completionEvidence) {
        this.completionEvidence = completionEvidence;
    }

    public Long getCompletedAt() {
        return completedAt;
    }

    public void setCompletedAt(Long completedAt) {
        this.completedAt = completedAt;
    }
}
