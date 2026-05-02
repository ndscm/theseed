#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

container_cli=${CONTAINER_CLI:-"docker"}

./build.sh

"${container_cli}" run --name "keycloak" \
  --interactive --rm --tty \
  --network "host" \
  --env KC_HOSTNAME="localhost" \
  --env KC_DB_URL="jdbc:postgresql://127.0.0.1/keycloak" \
  --env KC_DB_USERNAME \
  --env KC_DB_PASSWORD \
  --env KC_BOOTSTRAP_ADMIN_USERNAME="unsecure" \
  --env KC_BOOTSTRAP_ADMIN_PASSWORD="unsecure" \
  ghcr.io/ndscm/seed-cloud-keycloak-container:latest
