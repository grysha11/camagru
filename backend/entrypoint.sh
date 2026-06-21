#!/bin/sh
set -e

goose -dir ./sql/schema postgres "$DB_URL" up

exec "$@"