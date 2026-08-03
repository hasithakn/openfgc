#!/bin/sh
# Runs automatically on first boot (empty data dir) via docker-entrypoint-initdb.d.
# MYSQL_ROOT_PASSWORD is already exported by the official mysql entrypoint at this point.
# consent-server/dbscripts/db_schema_mysql.sql assumes a pre-selected database (no
# CREATE DATABASE/USE of its own), so this wraps it rather than needing that generated
# host-side — the real schema file is bind-mounted read-only at /schemas/consent-mgt-schema.sql.
set -e
mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "CREATE DATABASE IF NOT EXISTS consent_mgt;"
mysql -uroot -p"$MYSQL_ROOT_PASSWORD" consent_mgt < /schemas/consent-mgt-schema.sql
