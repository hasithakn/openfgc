package org.wso2.dpdp.enf.endpoint.exception;

import org.wso2.dpdp.enf.service.exception.ENFException;

import javax.ws.rs.core.MediaType;
import javax.ws.rs.core.Response;
import javax.ws.rs.ext.ExceptionMapper;
import javax.ws.rs.ext.Provider;
import java.util.UUID;
import java.util.logging.Level;
import java.util.logging.Logger;

@Provider
public class ENFExceptionMapper implements ExceptionMapper<Throwable> {

    private static final Logger LOGGER = Logger.getLogger(ENFExceptionMapper.class.getName());

    @Override
    public Response toResponse(Throwable exception) {
        if (exception instanceof ENFException enfEx) {
            ErrorEnvelope envelope = new ErrorEnvelope(
                    enfEx.getCode(),
                    enfEx.getMessage(),
                    enfEx.getDescription(),
                    UUID.randomUUID().toString()
            );
            return Response.status(enfEx.getStatusCode())
                    .type(MediaType.APPLICATION_JSON)
                    .entity(envelope)
                    .build();
        }

        LOGGER.log(Level.SEVERE, "Unhandled exception in ENF API: " + exception.getMessage(), exception);

        ErrorEnvelope envelope = new ErrorEnvelope(
                "CS-5000",
                "Internal error",
                "An unexpected error occurred while processing the request.",
                UUID.randomUUID().toString()
        );

        return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                .type(MediaType.APPLICATION_JSON)
                .entity(envelope)
                .build();
    }
}
