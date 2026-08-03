# Build context: event-framework/ — builds the shaded fat jar from source with Maven, no
# host Maven/JDK install needed. All config (DB type/url/user/pass, port) is env-var driven,
# no config file needed (see Main.java's initDatabase()/loadConfigFile()).
FROM maven:3.9-eclipse-temurin-17 AS builder
WORKDIR /src
COPY . .
RUN mvn -q -B clean package -DskipTests

FROM eclipse-temurin:17-jre
WORKDIR /app
COPY --from=builder /src/event-notification-runner/target/event-notification-runner-1.0.0-SNAPSHOT.jar /app/event-notification-runner.jar
EXPOSE 8082
# ENF has no dedicated health endpoint — a bare TCP connect (via bash's /dev/tcp) is the
# closest available proxy for "the Grizzly HTTP server has started listening".
HEALTHCHECK --interval=10s --timeout=5s --retries=10 --start-period=15s \
    CMD bash -c "echo > /dev/tcp/localhost/8082" || exit 1
ENTRYPOINT ["java", "-jar", "/app/event-notification-runner.jar"]
