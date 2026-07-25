package org.wso2.dpdp.enf.service;

import org.wso2.dpdp.enf.dao.model.Event;
import java.util.List;

public interface EventFanOutService {
    void fanOutEvent(Event event, List<String> eventPurposes);
}
