package org.wso2.dpdp.enf.dao;

import org.wso2.dpdp.enf.dao.model.Event;
import java.util.List;
import java.util.Optional;

public interface EventDAO {
    boolean addEvent(Event event);
    Optional<Event> getEventById(String eventId);
    void addEventPurposes(String eventId, List<String> purposes);
    List<String> getEventPurposes(String eventId);
    boolean hasActiveEventsForTopic(String topicId);
}
