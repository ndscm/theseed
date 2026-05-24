#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

container_engine=${CONTAINER_ENGINE:-"podman"}

./build.sh

"${container_engine}" run --name "sfe" --rm --interactive --tty \
  --network "host" \
  ghcr.io/ndscm/seed-cloud-sfe-container:latest \
  --http_port 8080 \
  --verbose
