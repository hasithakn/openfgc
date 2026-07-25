package org.wso2.dpdp.enf.endpoint;

import org.wso2.dpdp.enf.endpoint.bean.CompletionResponseBean;
import org.wso2.dpdp.enf.endpoint.bean.CompletionRequestBean;
import org.wso2.dpdp.enf.endpoint.handler.CompletionHandler;

import javax.ws.rs.Consumes;
import javax.ws.rs.HeaderParam;
import javax.ws.rs.POST;
import javax.ws.rs.Path;
import javax.ws.rs.PathParam;
import javax.ws.rs.Produces;
import javax.ws.rs.core.MediaType;
import javax.ws.rs.core.Response;

@Path("/deliveries/{deliveryId}/completion")
@Consumes(MediaType.APPLICATION_JSON)
@Produces(MediaType.APPLICATION_JSON)
public class CompletionEndpoint {

    private final CompletionHandler completionHandler;

    public CompletionEndpoint() {
        this.completionHandler = new CompletionHandler();
    }

    public CompletionEndpoint(CompletionHandler completionHandler) {
        this.completionHandler = completionHandler;
    }

    @POST
    public Response submitCompletion(
            @HeaderParam("X-Org-Id") String orgId,
            @HeaderParam("X-Group-Id") String groupId,
            @HeaderParam("X-Event-Signature") String signature,
            @PathParam("deliveryId") String deliveryId,
            CompletionRequestBean request) {

        String rawRequestBody = convertCompletionRequestToString(request);
        CompletionResponseBean response = completionHandler.submitCompletion(
                orgId, groupId, deliveryId, rawRequestBody, signature, request
        );

        return Response.status(Response.Status.CREATED).entity(response).build();
    }

    private String convertCompletionRequestToString(CompletionRequestBean request) {
        if (request == null) return "{}";
        StringBuilder sb = new StringBuilder("{");
        if (request.getCompletionStatus() != null) {
            sb.append("\"completionStatus\":\"").append(request.getCompletionStatus()).append("\"");
        }
        if (request.getCompletionEvidence() != null) {
            sb.append(",\"completionEvidence\":\"").append(request.getCompletionEvidence()).append("\"");
        }
        sb.append("}");
        return sb.toString();
    }
}
