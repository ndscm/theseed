#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

database_gateway="127.0.0.1"
database_name="golink"
database_login="golink"
database_secret_file="${ND_USER_SECRET_HOME}/golink/GOLINK_DATABASE_SECRET"
database_secret=$(cat "${database_secret_file}" | jq --raw-input --raw-output @uri)

npx @ariga/atlas schema inspect \
  --url "ent://schema" \
  --dev-url "postgres://${database_login}:${database_secret}@${database_gateway}:5432/${database_name}?search_path=migration&sslmode=disable" \
  --format '{{ sql . }}' >./schema.sql
