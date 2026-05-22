#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

container_engine=${CONTAINER_ENGINE:-"podman"}

./build.sh

"${container_engine}" run --name "stuff" --rm --interactive --tty \
  --network "host" \
  ghcr.io/ndscm/seed-office-stuff-container:latest
