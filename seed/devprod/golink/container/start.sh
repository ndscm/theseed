#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

container_engine=${CONTAINER_ENGINE:-"podman"}

export CONTAINER_ENGINE="${container_engine}"
./build.sh

"${container_engine}" run --name "golink" --rm --interactive --tty \
  --network "host" \
  ghcr.io/ndscm/seed-devprod-golink-container:latest
