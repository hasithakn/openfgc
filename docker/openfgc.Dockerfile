# Build context: consent-server/ (repo root's consent-server module) — builds the binary
# from source, no host Go toolchain needed. Alpine (not distroless) so `wget` is available
# for the HEALTHCHECK below.
FROM golang:1.26.3 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/consent-server ./cmd/server

FROM alpine:3.20
RUN addgroup -g 802 openfgc && adduser -u 802 -G openfgc -D -h /app openfgc
WORKDIR /app
COPY --from=builder /out/consent-server /app/consent-server
COPY cmd/server/repository/conf/deployment.docker.yaml /app/repository/conf/deployment.yaml
COPY dbscripts/ /app/dbscripts/
RUN chmod +x /app/consent-server && chown -R openfgc:openfgc /app
USER openfgc:openfgc
EXPOSE 8060
HEALTHCHECK --interval=10s --timeout=5s --retries=10 --start-period=10s \
    CMD wget -q -O /dev/null http://localhost:8060/health/liveness || exit 1
ENTRYPOINT ["/app/consent-server"]
