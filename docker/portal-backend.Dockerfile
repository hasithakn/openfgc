# Build context: demo-artifacts/portal-backend (populated by `demo.sh build` via `task build`)
# — just the prebuilt `bff` binary. Alpine (not distroless) so `wget` is available for the
# HEALTHCHECK below.
FROM alpine:3.20
RUN addgroup -g 802 bff && adduser -u 802 -G bff -D -h /app bff
WORKDIR /app
COPY bff /app/bff
RUN chmod +x /app/bff && chown -R bff:bff /app
USER bff:bff
EXPOSE 8081
HEALTHCHECK --interval=10s --timeout=5s --retries=10 --start-period=10s \
    CMD wget -q -O /dev/null http://localhost:8081/health/liveness || exit 1
ENTRYPOINT ["/app/bff"]
