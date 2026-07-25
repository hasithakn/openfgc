package org.wso2.dpdp.enf.endpoint;

import org.wso2.dpdp.enf.endpoint.bean.EventCreateRequestBean;
import org.wso2.dpdp.enf.endpoint.bean.EventResponseBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionEventHistoryResponseBean;
import org.wso2.dpdp.enf.endpoint.bean.SubscriptionEventListResponseBean;
import org.wso2.dpdp.enf.endpoint.handler.EventHandler;

import javax.ws.rs.Consumes;
import javax.ws.rs.GET;
import javax.ws.rs.HeaderParam;
import javax.ws.rs.POST;
import javax.ws.rs.Path;
import javax.ws.rs.PathParam;
import javax.ws.rs.Produces;
import javax.ws.rs.QueryParam;
import javax.ws.rs.core.MediaType;
import javax.ws.rs.core.Response;

@Path("/events")
@Consumes(MediaType.APPLICATION_JSON)
@Produces(MediaType.APPLICATION_JSON)
public class EventEndpoint {

    private final EventHandler eventHandler;

    public EventEndpoint() {
        this.eventHandler = new EventHandler();
    }

    public EventEndpoint(EventHandler eventHandler) {
        this.eventHandler = eventHandler;
    }

    @POST
    public Response publishEvent(
            @HeaderParam("X-Org-Id") String orgId,
            @HeaderParam("X-Group-Id") String groupId,
            EventCreateRequestBean request) {
        EventResponseBean response = eventHandler.publishEvent(orgId, groupId, request);
        return Response.status(Response.Status.CREATED).entity(response).build();
    }

    @GET
    public Response listOrgEvents(
            @HeaderParam("X-Org-Id") String orgId,
            @QueryParam("status") String status,
            @QueryParam("subscriptionId") String subscriptionId,
            @QueryParam("purposes") String purposes,
            @QueryParam("search") String search,
            @QueryParam("limit") Integer limit,
            @QueryParam("offset") Integer offset) {
        SubscriptionEventListResponseBean response = eventHandler.listOrgEvents(orgId, status, subscriptionId, purposes, search, limit, offset);
        return Response.ok(response).build();
    }

    @GET
    @Path("/{deliveryId}/history")
    public Response getOrgEventHistory(
            @HeaderParam("X-Org-Id") String orgId,
            @PathParam("deliveryId") String deliveryId) {
        SubscriptionEventHistoryResponseBean response = eventHandler.getOrgEventHistory(orgId, deliveryId);
        return Response.ok(response).build();
    }

    @GET
    @Path("/{deliveryId}")
    public Response getOrgEventHistoryAlias(
            @HeaderParam("X-Org-Id") String orgId,
            @PathParam("deliveryId") String deliveryId) {
        SubscriptionEventHistoryResponseBean response = eventHandler.getOrgEventHistory(orgId, deliveryId);
        return Response.ok(response).build();
    }
}
