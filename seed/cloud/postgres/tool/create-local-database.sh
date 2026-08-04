#!/usr/bin/env bash
set -eux
set -o pipefail

project_path="${1:-""}"
service="${2:-""}"

if [[ -z "${project_path}" || -z "${service}" ]]; then
  printf "Usage: %s <project_path> <service>\n" "${0}" >&2
  printf "Example: %s seed/newtype/steins steins\n" "${0}" >&2
  exit 1
fi

service_database_login="${service}_local"
service_database_secret_path="${project_path}/database/${service^^}_DATABASE_SECRET.local.age"

if [[ ! -f "$(ndscm secret --user get-path "${service_database_secret_path}")" ]]; then
  set +x
  service_database_secret="$(head -c 128 /dev/urandom | tr -dc 'A-Za-z0-9!@#' | head -c 16)"
  printf "%s" "${service_database_secret}" | ndscm secret --user encrypt "${service_database_secret_path}"
  set -x
fi

set +x
POSTGRES_USER="$(ndscm secret --user decrypt seed/cloud/postgres/POSTGRES_USER.local.age)"
POSTGRES_PASSWORD="$(ndscm secret --user decrypt seed/cloud/postgres/POSTGRES_PASSWORD.local.age)"
set -x

# TODO(nagi): run psql without docker

set +x
service_database_secret="$(ndscm secret --user decrypt "${service_database_secret_path}")"
docker exec postgres psql -U "${POSTGRES_USER}" -c "CREATE USER ${service_database_login} WITH PASSWORD '${service_database_secret}';" || true
set -x

docker exec postgres psql -U "${POSTGRES_USER}" -c "CREATE DATABASE ${service_database_login} OWNER ${service_database_login};" || true
docker exec postgres psql -U "${POSTGRES_USER}" -c "GRANT ALL PRIVILEGES ON DATABASE ${service_database_login} TO ${service_database_login};"
docker exec postgres psql -U "${service_database_login}" -d "${service_database_login}" -c "CREATE SCHEMA IF NOT EXISTS migration;"
