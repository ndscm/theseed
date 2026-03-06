#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

deployenv=${1:-""}

if [[ "${deployenv}" == "" ]]; then
  echo "Usage: $0 prod|local"
  exit 1
elif [[ "${deployenv}" == "prod" ]]; then
  jump="${JUMP}"
  database_gateway="${POSTGRES_PROD_HOST}"
  database_name="golink"
  database_login="golink"
  database_secret_file="${ND_MONOREPO_SECRET_HOME}/golink/golink/GOLINK_DATABASE_SECRET"
elif [[ "${deployenv}" == "local" ]]; then
  database_gateway="host.docker.internal"
  database_name="golink"
  database_login="golink"
  database_secret_file="${ND_USER_SECRET_HOME}/golink/GOLINK_DATABASE_SECRET"
else
  echo "Unknown deploy environment: ${deployenv}"
  exit 1
fi

bind=$(pwd)/seed/devprod/golink/database/schema.sql
if [[ -n "${jump:-}" ]]; then
  bazel run //devprod/docker:push_image -- ${jump} arigaio/atlas:latest
  export DOCKER_HOST="ssh://${jump}"
  scp "${bind}" "${jump}:/tmp/golink-schema.sql"
  bind="/tmp/golink-schema.sql"
fi

database_secret=$(cat "${database_secret_file}" | jq --raw-input --raw-output @uri)

docker run --rm --interactive \
  --add-host=host.docker.internal:host-gateway \
  --mount "type=bind,src=${bind},dst=/schema.sql" \
  arigaio/atlas:latest schema apply \
  --url "postgres://${database_login}:${database_secret}@${database_gateway}:5432/${database_name}?search_path=public&sslmode=disable" \
  --to "file:///schema.sql" \
  --dev-url "postgres://${database_login}:${database_secret}@${database_gateway}:5432/${database_name}?search_path=migration&sslmode=disable"
