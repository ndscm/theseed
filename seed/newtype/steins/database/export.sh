#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

database_gateway="127.0.0.1"
database_name="steins_local"
database_login="steins_local"
database_secret_path="seed/newtype/steins/database/STEINS_DATABASE_SECRET.local.age"

set +x
database_secret="$(ndscm secret --user decrypt "${database_secret_path}" | jq --raw-input --raw-output @uri)"
atlas schema inspect \
  --url "ent://schema" \
  --dev-url "postgres://${database_login}:${database_secret}@${database_gateway}:5432/${database_name}?search_path=migration&sslmode=disable" \
  --format '{{ sql . }}' >./schema.sql
set -x
