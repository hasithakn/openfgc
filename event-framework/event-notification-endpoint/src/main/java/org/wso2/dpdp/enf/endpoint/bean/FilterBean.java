package org.wso2.dpdp.enf.endpoint.bean;

import java.util.List;

public class FilterBean {

    private String type;
    private List<String> purposes;

    public FilterBean() {
    }

    public FilterBean(String type, List<String> purposes) {
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
