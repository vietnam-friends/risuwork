#!/bin/sh

set -ex
cd `dirname $0`

ISUCON_DB_HOST=${DB_HOST:-127.0.0.1}
ISUCON_DB_PORT=${DB_PORT:-3306}
ISUCON_DB_USER=${DB_USER:-isucon}
ISUCON_DB_PASS=${DB_PASS:-isucon}
ISUCON_DB_NAME=${DB_NAME:-risuwork}

# MySQLを初期化
mysql -u"$ISUCON_DB_USER" \
		-p"$ISUCON_DB_PASS" \
		--host "$ISUCON_DB_HOST" \
		--port "$ISUCON_DB_PORT" \
		"$ISUCON_DB_NAME" < 01_schema.sql


mysql -u"$ISUCON_DB_USER" \
		-p"$ISUCON_DB_PASS" \
		--host "$ISUCON_DB_HOST" \
		--port "$ISUCON_DB_PORT" \
		"$ISUCON_DB_NAME" < 02_testdata.sql
