#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

container_cli=${CONTAINER_CLI:-"docker"}

./build.sh

"${container_cli}" run --name "golink" --rm --interactive --tty \
  --network "host" \
  ghcr.io/ndscm/seed-devprod-golink-container:latest
