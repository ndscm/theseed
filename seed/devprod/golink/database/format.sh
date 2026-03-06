#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

database_gateway="host.docker.internal"
database_name="golink"
database_login="golink"
database_secret_file="${ND_USER_SECRET_HOME}/golink/GOLINK_DATABASE_SECRET"
database_secret=$(cat "${database_secret_file}" | jq --raw-input --raw-output @uri)

docker run --rm --interactive \
  --add-host=host.docker.internal:host-gateway \
  --mount "type=bind,src=$(pwd)/seed/devprod/golink/database/schema.sql,dst=/schema.sql" \
  arigaio/atlas:latest schema inspect \
  --url "file:///schema.sql" \
  --dev-url "postgres://${database_login}:${database_secret}@${database_gateway}:5432/${database_name}?search_path=migration&sslmode=disable" \
  --format '{{ sql . }}' >./seed/devprod/golink/database/schema.sql.tmp
mv ./seed/devprod/golink/database/schema.sql.tmp ./seed/devprod/golink/database/schema.sql
