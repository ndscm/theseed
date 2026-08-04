#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

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
  database_gateway="127.0.0.1"
  database_name="golink"
  database_login="golink"
  database_secret_file="${ND_USER_SECRET_HOME}/golink/GOLINK_DATABASE_SECRET"
else
  echo "Unknown deploy environment: ${deployenv}"
  exit 1
fi

schema="$(pwd)/seed/devprod/golink/database/schema.sql"

if [[ ! -f "${schema}" ]]; then
  echo "Schema file not found: ${schema}"
  echo "Run seed/devprod/golink/database/export.sh to generate it."
  exit 1
fi

database_port=5432

if [[ -n "${jump:-}" ]]; then
  # Atlas runs locally now, so reach the prod database through an SSH tunnel
  # on the jump host and point atlas at the local end of the forward. Open a
  # master connection, then ask it to forward a dynamically allocated local
  # port (":0") to the database; ssh prints the port it picked.
  control_socket="$(mktemp -u)"
  ssh -o ExitOnForwardFailure=yes -fN -M -S "${control_socket}" "${jump}"
  trap 'ssh -S "${control_socket}" -O exit "${jump}" >/dev/null 2>&1 || true' EXIT
  local_port="$(ssh -S "${control_socket}" -O forward -L "0:${database_gateway}:5432" "${jump}")"
  database_gateway="127.0.0.1"
  database_port="${local_port}"
fi

set +x
database_secret=$(cat "${database_secret_file}" | jq --raw-input --raw-output @uri)
atlas schema apply \
  --url "postgres://${database_login}:${database_secret}@${database_gateway}:${database_port}/${database_name}?search_path=public&sslmode=disable" \
  --to "file://${schema}" \
  --dev-url "postgres://${database_login}:${database_secret}@${database_gateway}:${database_port}/${database_name}?search_path=migration&sslmode=disable"
set -x
