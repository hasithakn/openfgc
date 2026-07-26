package org.wso2.dpdp.enf.endpoint;

import org.wso2.dpdp.enf.endpoint.bean.PollRequestBean;
import org.wso2.dpdp.enf.endpoint.bean.PollResponseBean;
import org.wso2.dpdp.enf.endpoint.handler.PollHandler;

import javax.ws.rs.Consumes;
import javax.ws.rs.HeaderParam;
import javax.ws.rs.POST;
import javax.ws.rs.Path;
import javax.ws.rs.Produces;
import javax.ws.rs.core.MediaType;
import javax.ws.rs.core.Response;

@Path("/events/poll")
@Consumes(MediaType.APPLICATION_JSON)
@Produces(MediaType.APPLICATION_JSON)
public class PollEndpoint {

    private final PollHandler pollHandler;

    public PollEndpoint() {
        this.pollHandler = new PollHandler();
    }

    public PollEndpoint(PollHandler pollHandler) {
        this.pollHandler = pollHandler;
    }

    @POST
    public Response pollEvents(
            @HeaderParam("org-id") String orgId,
            @HeaderParam("group-id") String groupId,
            @HeaderParam("X-Event-Signature") String signature,
            PollRequestBean request) {
        // String representation of request body for HMAC verification
        String rawRequestBody = convertPollRequestToString(request);
        PollResponseBean response = pollHandler.pollEvents(orgId, groupId, rawRequestBody, signature, request);
        return Response.ok(response).build();
    }

    private String convertPollRequestToString(PollRequestBean request) {
        if (request == null) return "{}";
        StringBuilder sb = new StringBuilder("{");
        sb.append("\"returnImmediately\":").append(Boolean.TRUE.equals(request.getReturnImmediately()));
        if (request.getMaxEvents() != null) {
            sb.append(",\"maxEvents\":").append(request.getMaxEvents());
        }
        sb.append("}");
        return sb.toString();
    }
}
