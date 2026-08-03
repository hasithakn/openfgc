# Build context: portal/backend/ — builds the BFF binary from source, no host Go toolchain
# needed. Alpine (not distroless) so `wget` is available for the HEALTHCHECK below.
FROM golang:1.26.3 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/bff ./cmd/server

FROM alpine:3.20
RUN addgroup -g 802 bff && adduser -u 802 -G bff -D -h /app bff
WORKDIR /app
COPY --from=builder /out/bff /app/bff
RUN chmod +x /app/bff && chown -R bff:bff /app
USER bff:bff
EXPOSE 8081
HEALTHCHECK --interval=10s --timeout=5s --retries=10 --start-period=10s \
    CMD wget -q -O /dev/null http://localhost:8081/health/liveness || exit 1
ENTRYPOINT ["/app/bff"]
