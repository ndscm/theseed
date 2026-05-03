#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

container_cli=${CONTAINER_CLI:-"docker"}

./build.sh

"${container_cli}" run --name "sfe" \
  --interactive --rm --tty \
  --network "host" \
  ghcr.io/ndscm/seed-cloud-sfe-container:latest
