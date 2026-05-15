#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

user="${1:-"jenkins"}"

container_engine=${CONTAINER_ENGINE:-"podman"}
if [[ "${container_engine}" != "podman" ]]; then
  printf 'Only podman is supported for now\n'
  exit 1
fi

./build.sh

podman run --name jenkins-node \
  --interactive --rm --tty \
  -u "${user}" \
  ghcr.io/ndscm/seed-devprod-jenkins-node-container:latest \
  /bin/bash
