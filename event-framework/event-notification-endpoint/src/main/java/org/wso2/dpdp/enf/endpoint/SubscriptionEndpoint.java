package org.wso2.dpdp.enf.endpoint;

import org.wso2.dpdp.enf.endpoint.bean.SubscriptionCreateRequestBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionEventHistoryResponseBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionEventListResponseBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionListResponseBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionResponseBean;
import org.wso2.dpdp.enf.endpoint.handler.SubscriptionHandler;

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

@Path("/subscriptions")
@Consumes(MediaType.APPLICATION_JSON)
@Produces(MediaType.APPLICATION_JSON)
public class SubscriptionEndpoint {

    private final SubscriptionHandler subscriptionHandler;

    public SubscriptionEndpoint() {
        this.subscriptionHandler = new SubscriptionHandler();
    }

    public SubscriptionEndpoint(SubscriptionHandler subscriptionHandler) {
        this.subscriptionHandler = subscriptionHandler;
    }

    @POST
    public Response createSubscription(
            @HeaderParam("X-Org-Id") String orgId,
            @HeaderParam("X-Group-Id") String groupId,
            SubscriptionCreateRequestBean request) {
        SubscriptionResponseBean response = subscriptionHandler.createSubscription(orgId, groupId, request);
        if (Boolean.TRUE.equals(response.getAlreadyExists())) {
            return Response.ok(response).build();
        }
        return Response.status(Response.Status.CREATED).entity(response).build();
    }

    @GET
    public Response listSubscriptions(
            @HeaderParam("X-Org-Id") String orgId,
            @QueryParam("status") String status,
            @QueryParam("purposes") String purposes,
            @QueryParam("search") String search,
            @QueryParam("limit") Integer limit,
            @QueryParam("offset") Integer offset,
            @QueryParam("sort") String sort) {
        SubscriptionListResponseBean response = subscriptionHandler.listSubscriptions(orgId, status, purposes, search, limit, offset, sort);
        return Response.ok(response).build();
    }

    @GET
    @Path("/{subscriptionId}")
    public Response getSubscription(
            @HeaderParam("X-Org-Id") String orgId,
            @PathParam("subscriptionId") String subscriptionId) {
        SubscriptionResponseBean response = subscriptionHandler.getSubscription(orgId, subscriptionId);
        return Response.ok(response).build();
    }

    @DELETE
    @Path("/{subscriptionId}")
    public Response deleteSubscription(
            @HeaderParam("X-Org-Id") String orgId,
            @PathParam("subscriptionId") String subscriptionId) {
        subscriptionHandler.deleteSubscription(orgId, subscriptionId);
        return Response.noContent().build();
    }

    @GET
    @Path("/{subscriptionId}/events")
    public Response listSubscriptionEvents(
            @HeaderParam("X-Org-Id") String orgId,
            @PathParam("subscriptionId") String subscriptionId,
            @QueryParam("limit") Integer limit,
            @QueryParam("offset") Integer offset) {
        SubscriptionEventListResponseBean response = subscriptionHandler.listSubscriptionEvents(orgId, subscriptionId, limit, offset);
        return Response.ok(response).build();
    }

    @GET
    @Path("/{subscriptionId}/events/{deliveryId}/history")
    public Response getSubscriptionEventHistory(
            @HeaderParam("X-Org-Id") String orgId,
            @PathParam("subscriptionId") String subscriptionId,
            @PathParam("deliveryId") String deliveryId) {
        SubscriptionEventHistoryResponseBean response = subscriptionHandler.getSubscriptionEventHistory(orgId, subscriptionId, deliveryId);
        return Response.ok(response).build();
    }

    @GET
    @Path("/{subscriptionId}/events/{deliveryId}")
    public Response getSubscriptionEventHistoryAlias(
            @HeaderParam("X-Org-Id") String orgId,
            @PathParam("subscriptionId") String subscriptionId,
            @PathParam("deliveryId") String deliveryId) {
        SubscriptionEventHistoryResponseBean response = subscriptionHandler.getSubscriptionEventHistory(orgId, subscriptionId, deliveryId);
        return Response.ok(response).build();
    }
}
