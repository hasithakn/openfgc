#!/bin/sh
# Runs automatically on first boot (empty data dir) via docker-entrypoint-initdb.d.
# event-framework's mysql.sql assumes a pre-selected database the same way consent-server's
# does — see 01-consent-mgt.sh's comment. Real schema file bind-mounted read-only at
# /schemas/enf-schema.sql.
set -e
mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "CREATE DATABASE IF NOT EXISTS enf_db;"
mysql -uroot -p"$MYSQL_ROOT_PASSWORD" enf_db < /schemas/enf-schema.sql
