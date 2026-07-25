package org.wso2.dpdp.enf.service.dto;

import java.util.List;

public class FilterDTO {

    private String type;
    private List<String> purposes;

    public FilterDTO() {
    }

    public FilterDTO(String type, List<String> purposes) {
        this.type = type;
        this.purposes = purposes;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }

    public List<String> getPurposes() {
        return purposes;
    }

    public void setPurposes(List<String> purposes) {
        this.purposes = purposes;
    }
}
