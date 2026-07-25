package org.wso2.dpdp.enf.service.matching;

import java.util.List;

public class FilterMatcher {

    private FilterMatcher() {
    }

    /**
     * Matches event purposes against subscription purpose filter rule.
     *
     * @param filterMode          "all", "all_except", or "specific"
     * @param subscriptionPurposes Purposes configured on the subscription
     * @param eventPurposes        Purposes attached to the published event
     * @return true if the subscription matches the event, false otherwise
     */
    public static boolean isMatch(String filterMode, List<String> subscriptionPurposes, List<String> eventPurposes) {
        if (filterMode == null) {
            return false;
        }

        switch (filterMode.toLowerCase()) {
            case "all":
                return true;

            case "all_except":
                if (eventPurposes == null || eventPurposes.isEmpty()) {
                    return true;
                }
                if (subscriptionPurposes == null || subscriptionPurposes.isEmpty()) {
                    return true;
                }
                // Return true if NONE of event's purposes are in subscriptionPurposes
                for (String ep : eventPurposes) {
                    if (subscriptionPurposes.stream().anyMatch(sp -> sp.equalsIgnoreCase(ep))) {
                        return false;
                    }
                }
                return true;

            case "specific":
                if (eventPurposes == null || eventPurposes.isEmpty()) {
                    return false;
                }
                if (subscriptionPurposes == null || subscriptionPurposes.isEmpty()) {
                    return false;
                }
                // Return true if AT LEAST ONE of event's purposes is in subscriptionPurposes
                for (String ep : eventPurposes) {
                    if (subscriptionPurposes.stream().anyMatch(sp -> sp.equalsIgnoreCase(ep))) {
                        return true;
                    }
                }
                return false;

            default:
                return false;
        }
    }
}
