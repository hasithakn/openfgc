# Build context: demo-artifacts/openfgc (populated by `demo.sh build`) — a prebuilt static
# Go binary + config + dbscripts, mirroring consent-server's own target/server/ layout.
# Alpine (not distroless) so `wget` is available for the HEALTHCHECK below.
FROM alpine:3.20
RUN addgroup -g 802 openfgc && adduser -u 802 -G openfgc -D -h /app openfgc
WORKDIR /app
COPY . /app
RUN chmod +x /app/consent-server && chown -R openfgc:openfgc /app
USER openfgc:openfgc
EXPOSE 8060
HEALTHCHECK --interval=10s --timeout=5s --retries=10 --start-period=10s \
    CMD wget -q -O /dev/null http://localhost:8060/health/liveness || exit 1
ENTRYPOINT ["/app/consent-server"]
