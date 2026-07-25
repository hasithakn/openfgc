package org.wso2.dpdp.enf.dao.constants;

public class DAOConstants {

    private DAOConstants() {
    }

    // Table Names
    public static final String TABLE_TOPIC = "TOPIC";
    public static final String TABLE_EVENT = "EVENT";
    public static final String TABLE_EVENT_PURPOSE = "EVENT_PURPOSE";
    public static final String TABLE_SUBSCRIPTION = "SUBSCRIPTION";
    public static final String TABLE_SUBSCRIPTION_PURPOSE = "SUBSCRIPTION_PURPOSE";
    public static final String TABLE_WEBHOOK_DELIVERY = "WEBHOOK_DELIVERY";
    public static final String TABLE_WEBHOOK_DELIVERY_ACK = "WEBHOOK_DELIVERY_ACK";
    public static final String TABLE_WEBHOOK_DELIVERY_AUDIT = "WEBHOOK_DELIVERY_AUDIT";
    public static final String TABLE_POLL_DELIVERY = "POLL_DELIVERY";

    // Status Enums / Constants
    public static final String TOPIC_STATUS_ACTIVE = "active";
    public static final String TOPIC_STATUS_DEREGISTERED = "deregistered";

    public static final String SUB_STATUS_ACTIVE = "active";
    public static final String SUB_STATUS_STALE = "stale";
    public static final String SUB_STATUS_DELETED = "deleted";

    public static final String PURPOSE_FILTER_ALL = "all";
    public static final String PURPOSE_FILTER_ALL_EXCEPT = "all_except";
    public static final String PURPOSE_FILTER_SPECIFIC = "specific";

    public static final String DELIVERY_MODE_WEBHOOK = "webhook";
    public static final String DELIVERY_MODE_PULL = "pull";
    public static final String DELIVERY_MODE_POLL = "poll";

    public static final String WEBHOOK_STATUS_PENDING = "pending";
    public static final String WEBHOOK_STATUS_DELIVERED = "delivered";
    public static final String WEBHOOK_STATUS_COMPLETED = "completed";
    public static final String WEBHOOK_STATUS_FAILED = "failed";

    public static final String ACK_STATUS_COMPLETED = "completed";
    public static final String ACK_STATUS_DISPUTED = "disputed";
    public static final String ACK_STATUS_PARTIAL = "partial";

    public static final String POLL_STATUS_PENDING = "pending";
    public static final String POLL_STATUS_ACKNOWLEDGED = "acknowledged";
    public static final String POLL_STATUS_ERR = "err";
}
