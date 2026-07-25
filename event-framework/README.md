# Event Notification Framework

This repository contains the ENF Java services and a simple Node.js webhook echo listener for local testing.

## Prerequisites
- Java 17
- Maven 3.6+
- Node.js (18+ recommended) — only required to run the sample webhook listener
- Postman (to import `ENF_Postman_Collection.json`)

DB
mysql -uroot -ppassword temp-demo-dpdp-portal-event < event-framework/event-notification-dao/src/main/resources/dbscripts/mysql.sql

## Build the Java services
1. From the repository root, build all modules:

```bash
mvn clean install -DskipTests
```

2. Run the runner module (starts the ENF HTTP server on port 8080):

```bash
mvn -pl event-notification-runner exec:java
```

Notes:
- The runner's main class is `org.wso2.dpdp.enf.runner.Main` and it binds to `http://0.0.0.0:8080/`.
- The application uses an in-memory H2 database by default. To use MySQL, set environment variables or system properties (examples below).

Supported DB environment variables (optional):
- `ENF_DB_TYPE` (use `mysql` to enable MySQL)
- `ENF_DB_URL`, `ENF_DB_USER`, `ENF_DB_PASS`

Example (run with MySQL parameters):

```bash
# set env vars in the shell, then run the runner
export ENF_DB_TYPE=mysql
export ENF_DB_URL="jdbc:mysql://localhost:3306/enf_db?createDatabaseIfNotExist=true&useSSL=false"
export ENF_DB_USER=root
export ENF_DB_PASS=root
mvn -pl event-notification-runner exec:java
```

## Run the webhook echo listener (Node)
This script listens on port 9091 by default and echoes verification challenges or logs received payloads.

1. Start the listener:

```bash
# from repository root
node webhook-listener.js
```

2. To change the port, set `PORT` environment variable, e.g. `PORT=3000 node webhook-listener.js`.

## Postman: Import and test
1. Import the included `ENF_Postman_Collection.json` into Postman.
2. Verify collection variables (`baseUrl`, `orgId`, `groupId`). By default `baseUrl` is `http://localhost:8080`.

Basic test workflow:

1. Start the Java runner (see above).
2. Start the Node webhook listener on port 9091.
3. In Postman, create a subscription that uses a webhook delivery callback. Example request body (modify the `Create Subscription` request or send a new POST to `{{baseUrl}}/subscriptions`):

```json
{
  "topic": "user-consent-events",
  "filter": { "expression": "purpose == 'MARKETING'" },
  "delivery": {
    "mode": "WEBHOOK",
    "callbackUrl": "http://localhost:9091/webhook"
  }
}
```

4. Publish an event via Postman (`Publish Event` request). If the subscription matches, ENF will attempt to POST the event to the webhook listener.
5. Observe the webhook listener console — it will log the incoming request and echo challenge parameters if present.

Testing webhook verification manually (Postman/curl):

- GET challenge (should echo the challenge token):

```bash
curl "http://localhost:9091/webhook?challenge=abc123"
# response: abc123
```

- POST JSON body with challenge (should return JSON with challenge and status):

```bash
curl -X POST http://localhost:9091/webhook -H "Content-Type: application/json" -d '{"challenge":"abc123"}'
# response: {"challenge":"abc123","status":"verified"}
```

## Troubleshooting
- If Postman requests to `{{baseUrl}}` return connection errors, ensure the runner is running and port 8080 is free.
- If the webhook listener is not receiving callbacks, ensure the subscription's `callbackUrl` is correct and reachable from the machine running the runner.
- Use the runner logs to inspect delivery attempts and failures.

## Postman collection
- The repository includes `ENF_Postman_Collection.json` pre-configured with example requests for topics, subscriptions, events, and delivery completion.

## Next steps
- Add automated integration tests (optional).
