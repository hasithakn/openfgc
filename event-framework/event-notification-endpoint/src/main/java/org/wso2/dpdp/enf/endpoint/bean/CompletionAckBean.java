package org.wso2.dpdp.enf.endpoint.bean;

/**
 * @deprecated Use {@link CompletionResponseBean} instead.
 */
@Deprecated
public class CompletionAckBean extends CompletionResponseBean {

    public CompletionAckBean() {
        super();
    }

    public CompletionAckBean(String ackId, String deliveryId, String completionStatus, Long completedAt) {
        super(ackId, deliveryId, completionStatus, completedAt);
    }
}
