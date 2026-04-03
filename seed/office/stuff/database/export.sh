#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

database_gateway="127.0.0.1"
database_name="stuff"
database_login="stuff"
database_secret_file="${ND_USER_SECRET_HOME}/stuff/STUFF_DATABASE_SECRET"
set +x
database_secret=$(cat "${database_secret_file}" | jq --raw-input --raw-output @uri)
set -x

set +x
npx @ariga/atlas schema inspect \
  --url "ent://schema" \
  --dev-url "postgres://${database_login}:${database_secret}@${database_gateway}:5432/${database_name}?search_path=migration&sslmode=disable" \
  --format '{{ sql . }}' >./schema.sql
set -x
