#!/bin/sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$SCRIPT_DIR/../.."
BACKEND_DIR="$SCRIPT_DIR/.."

cd "$REPO_ROOT"
set -a
. ./.env
set +a

docker compose -f docker-compose.test.yaml up -d db-test

echo "Waiting for db-test to be ready..."
until docker compose -f docker-compose.test.yaml exec -T db-test pg_isready -U "$POSTGRES_USER" >/dev/null 2>&1; do
  sleep 1
done

if command -v goose >/dev/null 2>&1; then
  GOOSE="goose"
else
  GOOSE="go run github.com/pressly/goose/v3/cmd/goose@latest"
fi

cd "$BACKEND_DIR"
$GOOSE -dir ./sql/schema postgres "$TEST_DB_URL" up

TEST_DB_URL="$TEST_DB_URL" go test -p 1 ./...
