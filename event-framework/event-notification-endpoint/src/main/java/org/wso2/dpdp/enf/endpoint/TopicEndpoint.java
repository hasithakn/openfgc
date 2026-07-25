package org.wso2.dpdp.enf.endpoint;

import org.wso2.dpdp.enf.endpoint.bean.TopicCreateRequestBean;
import org.wso2.dpdp.enf.endpoint.bean.TopicListResponseBean;
import org.wso2.dpdp.enf.endpoint.bean.TopicResponseBean;
import org.wso2.dpdp.enf.endpoint.handler.TopicHandler;

import javax.ws.rs.Consumes;
import javax.ws.rs.DELETE;
import javax.ws.rs.GET;
import javax.ws.rs.HeaderParam;
import javax.ws.rs.POST;
import javax.ws.rs.Path;
import javax.ws.rs.PathParam;
import javax.ws.rs.Produces;
import javax.ws.rs.QueryParam;
import javax.ws.rs.core.MediaType;
import javax.ws.rs.core.Response;

@Path("/topics")
@Consumes(MediaType.APPLICATION_JSON)
@Produces(MediaType.APPLICATION_JSON)
public class TopicEndpoint {

    private final TopicHandler topicHandler;

    public TopicEndpoint() {
        this.topicHandler = new TopicHandler();
    }

    public TopicEndpoint(TopicHandler topicHandler) {
        this.topicHandler = topicHandler;
    }

    @POST
    public Response createTopic(@HeaderParam("X-Org-Id") String orgId, TopicCreateRequestBean request) {
        TopicResponseBean response = topicHandler.createTopic(orgId, request);
        return Response.status(Response.Status.CREATED).entity(response).build();
    }

    @GET
    public Response listTopics(
            @HeaderParam("X-Org-Id") String orgId,
            @QueryParam("status") String status,
            @QueryParam("search") String search,
            @QueryParam("limit") Integer limit,
            @QueryParam("offset") Integer offset,
            @QueryParam("count") Integer count,
            @QueryParam("sort") String sort) {
        TopicListResponseBean response = topicHandler.listTopics(orgId, status, search, limit, offset, count, sort);
        return Response.ok(response).build();
    }

    @DELETE
    @Path("/{topicId}")
    public Response deleteTopic(
            @HeaderParam("X-Org-Id") String orgId,
            @PathParam("topicId") String topicId) {
        TopicResponseBean response = topicHandler.deleteTopic(orgId, topicId);
        return Response.ok(response).build();
    }
}
