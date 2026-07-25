package org.wso2.dpdp.enf.service.dto;

import java.util.List;

public class PollResponseDTO {

    private List<PollDeliveryItemDTO> events;

    public PollResponseDTO() {
    }

    public PollResponseDTO(List<PollDeliveryItemDTO> events) {
        this.events = events;
    }

    public List<PollDeliveryItemDTO> getEvents() {
        return events;
    }

    public void setEvents(List<PollDeliveryItemDTO> events) {
        this.events = events;
    }
}
