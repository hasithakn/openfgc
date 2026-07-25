package org.wso2.dpdp.enf.runner;

import org.glassfish.grizzly.http.server.HttpServer;
import org.glassfish.jersey.grizzly2.httpserver.GrizzlyHttpServerFactory;
import org.glassfish.jersey.jackson.JacksonFeature;
import org.glassfish.jersey.server.ResourceConfig;
import org.wso2.dpdp.enf.dao.util.DBUtil;
import org.wso2.dpdp.enf.endpoint.CompletionEndpoint;
import org.wso2.dpdp.enf.endpoint.EventEndpoint;
import org.wso2.dpdp.enf.endpoint.PollEndpoint;
import org.wso2.dpdp.enf.endpoint.SubscriptionEndpoint;
import org.wso2.dpdp.enf.endpoint.TopicEndpoint;
import org.wso2.dpdp.enf.endpoint.exception.ENFExceptionMapper;

import org.wso2.dpdp.enf.service.impl.WebhookDeliveryProcessor;

import java.io.BufferedReader;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.net.URI;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.Statement;
import java.util.logging.Level;
import java.util.logging.Logger;

public class Main {

    private static final Logger LOGGER = Logger.getLogger(Main.class.getName());

    public static final String BASE_URI = "http://0.0.0.0:8080/";

    public static void main(String[] args) {
        try {
            LOGGER.info("Starting ENF Database Initialization...");
            initDatabase();

            LOGGER.info("Starting Grizzly HTTP Server on " + BASE_URI + "...");
            ResourceConfig config = new ResourceConfig()
                    .register(TopicEndpoint.class)
                    .register(SubscriptionEndpoint.class)
                    .register(EventEndpoint.class)
                    .register(PollEndpoint.class)
                    .register(CompletionEndpoint.class)
                    .register(ENFExceptionMapper.class)
                    .register(JacksonFeature.class);

            HttpServer server = GrizzlyHttpServerFactory.createHttpServer(URI.create(BASE_URI), config);
            server.start();

            // Start Webhook Delivery Retry Scheduled Job (runs every 5s for pending/failed retries)
            WebhookDeliveryProcessor.getInstance().startScheduledJob();

            LOGGER.info("==========================================================");
            LOGGER.info("  WSO2 DPDP Event Notification Framework (ENF) Running!   ");
            LOGGER.info("  Base URL: http://localhost:8080/                       ");
            LOGGER.info("  Ready to accept Postman requests. Press Ctrl+C to stop.  ");
            LOGGER.info("==========================================================");

            Thread.currentThread().join();
        } catch (Exception e) {
            LOGGER.log(Level.SEVERE, "Failed to start ENF HTTP server", e);
        }
    }

    private static void initDatabase() {
        loadConfigFile();

        String dbType = System.getProperty("ENF_DB_TYPE", System.getenv("ENF_DB_TYPE"));
        if (dbType == null) {
            dbType = System.getProperty("DB_TYPE", System.getenv("DB_TYPE"));
        }

        String defaultUrl = "jdbc:h2:mem:enf_db;DB_CLOSE_DELAY=-1;MODE=MySQL";
        String defaultUser = "sa";
        String defaultPass = "";

        if ("mysql".equalsIgnoreCase(dbType)) {
            defaultUrl = "jdbc:mysql://localhost:3306/enf_db?createDatabaseIfNotExist=true&useSSL=false&allowPublicKeyRetrieval=true";
            defaultUser = "root";
            defaultPass = "root";
        }

        String dbUrl = System.getProperty("ENF_DB_URL", System.getenv("ENF_DB_URL") != null ? System.getenv("ENF_DB_URL") : defaultUrl);
        String dbUser = System.getProperty("ENF_DB_USER", System.getenv("ENF_DB_USER") != null ? System.getenv("ENF_DB_USER") : defaultUser);
        String dbPass = System.getProperty("ENF_DB_PASS", System.getenv("ENF_DB_PASS") != null ? System.getenv("ENF_DB_PASS") : defaultPass);

        System.setProperty("ENF_DB_URL", dbUrl);
        System.setProperty("ENF_DB_USER", dbUser);
        System.setProperty("ENF_DB_PASS", dbPass);

        boolean isMysql = dbUrl.startsWith("jdbc:mysql:");

        try (Connection conn = DriverManager.getConnection(dbUrl, dbUser, dbPass);
             Statement stmt = conn.createStatement()) {

            LOGGER.info("Connected to database (" + (isMysql ? "MySQL" : "H2") + "): " + dbUrl);

            InputStream is = Main.class.getClassLoader().getResourceAsStream("dbscripts/mysql.sql");
            if (is == null) {
                is = DBUtil.class.getClassLoader().getResourceAsStream("dbscripts/mysql.sql");
            }

            if (is != null) {
                StringBuilder sqlBuilder = new StringBuilder();
                try (BufferedReader reader = new BufferedReader(new InputStreamReader(is))) {
                    String line;
                    while ((line = reader.readLine()) != null) {
                        line = line.trim();
                        if (line.startsWith("--") || line.isEmpty()) continue;
                        sqlBuilder.append(line).append(" ");
                    }
                }

                String[] statements = sqlBuilder.toString().split(";");
                for (String sql : statements) {
                    String statementToExecute = isMysql ? sql.trim() : sanitizeSqlForH2(sql.trim());
                    if (!statementToExecute.isEmpty()) {
                        try {
                            stmt.execute(statementToExecute);
                        } catch (Exception ex) {
                            LOGGER.warning("Warning executing DDL statement [" + statementToExecute + "]: " + ex.getMessage());
                        }
                    }
                }

                // Schema migration for existing MySQL tables
                if (isMysql) {
                    try {
                        stmt.execute("ALTER TABLE SUBSCRIPTION MODIFY COLUMN PURPOSE_FILTER_MODE varchar(32) NOT NULL");
                    } catch (Exception ignored) {}
                    try {
                        stmt.execute("ALTER TABLE SUBSCRIPTION DROP CHECK CHK_S_STATUS");
                    } catch (Exception ignored) {}
                    try {
                        stmt.execute("ALTER TABLE SUBSCRIPTION ADD CONSTRAINT CHK_S_STATUS CHECK (STATUS in ('active', 'pending', 'stale', 'deleted'))");
                    } catch (Exception ignored) {}
                    try {
                        stmt.execute("ALTER TABLE WEBHOOK_DELIVERY_ACK DROP CHECK CHK_EDA_COMPLETION_STATUS");
                    } catch (Exception ignored) {}
                    try {
                        stmt.execute("ALTER TABLE WEBHOOK_DELIVERY_ACK ADD CONSTRAINT CHK_EDA_COMPLETION_STATUS CHECK (COMPLETION_STATUS in ('completed', 'ack', 'disputed', 'partial'))");
                    } catch (Exception ignored) {}
                }

                LOGGER.info("Database schema initialized successfully.");
            } else {
                LOGGER.warning("mysql.sql script not found on classpath, skipping DDL execution.");
            }
        } catch (Exception e) {
            LOGGER.log(Level.WARNING, "Error initializing database tables", e);
        }
    }

    private static void loadConfigFile() {
        String[] configFiles = {"config.properties", "config.toml", "config.env"};
        for (String fileName : configFiles) {
            java.io.File file = new java.io.File(fileName);
            if (!file.exists()) {
                file = new java.io.File("event-notification-runner/" + fileName);
            }
            if (file.exists() && file.isFile()) {
                LOGGER.info("Loading configuration from: " + file.getAbsolutePath());
                try (BufferedReader reader = new BufferedReader(new java.io.FileReader(file))) {
                    String line;
                    while ((line = reader.readLine()) != null) {
                        line = line.trim();
                        if (line.isEmpty() || line.startsWith("#") || line.startsWith("//")) continue;
                        if (line.contains("=")) {
                            String[] parts = line.split("=", 2);
                            String key = parts[0].trim().replaceAll("^[\"'\\s]+|[\"'\\s]+$", "");
                            String val = parts[1].trim().replaceAll("^[\"'\\s]+|[\"'\\s]+$", "");

                            if ("db.type".equalsIgnoreCase(key) || "type".equalsIgnoreCase(key) || "ENF_DB_TYPE".equalsIgnoreCase(key)) {
                                if (System.getProperty("ENF_DB_TYPE") == null) System.setProperty("ENF_DB_TYPE", val);
                            } else if ("db.url".equalsIgnoreCase(key) || "url".equalsIgnoreCase(key) || "ENF_DB_URL".equalsIgnoreCase(key)) {
                                if (System.getProperty("ENF_DB_URL") == null) System.setProperty("ENF_DB_URL", val);
                            } else if ("db.user".equalsIgnoreCase(key) || "user".equalsIgnoreCase(key) || "ENF_DB_USER".equalsIgnoreCase(key)) {
                                if (System.getProperty("ENF_DB_USER") == null) System.setProperty("ENF_DB_USER", val);
                            } else if ("db.password".equalsIgnoreCase(key) || "db.pass".equalsIgnoreCase(key) || "password".equalsIgnoreCase(key) || "pass".equalsIgnoreCase(key) || "ENF_DB_PASS".equalsIgnoreCase(key)) {
                                if (System.getProperty("ENF_DB_PASS") == null) System.setProperty("ENF_DB_PASS", val);
                            }
                        }
                    }
                } catch (Exception e) {
                    LOGGER.warning("Could not read configuration file " + fileName + ": " + e.getMessage());
                }
                break;
            }
        }
    }

    private static String sanitizeSqlForH2(String sql) {
        if (sql == null) return "";
        return sql.replaceAll("(?i)ENGINE\\s*=\\s*InnoDB", "")
                  .replaceAll("(?i)DEFAULT\\s+CHARSET\\s*=\\s*utf8mb4", "")
                  .replaceAll("(?i)COLLATE\\s*=?\\s*utf8mb4_[^\\s,;()]+", "")
                  .replaceAll("(?i)CHARACTER\\s+SET\\s+utf8mb4", "")
                  .replaceAll("(?i)ON\\s+UPDATE\\s+CURRENT_TIMESTAMP", "")
                  .replaceAll("_utf8mb4'", "'")
                  .trim();
    }
}
