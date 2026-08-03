# Build context: demo-artifacts/event-framework (populated by `demo.sh build`) — just the
# shaded fat jar. All config (DB type/url/user/pass, port) is env-var driven, no config file
# needed (see event-framework Main.java's initDatabase()/loadConfigFile()).
# Ubuntu-based (not the alpine tag) — eclipse-temurin only publishes alpine variants for amd64,
# and this needs to build on arm64 hosts too.
FROM eclipse-temurin:17-jre
WORKDIR /app
COPY event-notification-runner.jar /app/event-notification-runner.jar
EXPOSE 8082
# ENF has no dedicated health endpoint — a bare TCP connect (via bash's /dev/tcp) is the
# closest available proxy for "the Grizzly HTTP server has started listening".
HEALTHCHECK --interval=10s --timeout=5s --retries=10 --start-period=15s \
    CMD bash -c "echo > /dev/tcp/localhost/8082" || exit 1
ENTRYPOINT ["java", "-jar", "/app/event-notification-runner.jar"]
