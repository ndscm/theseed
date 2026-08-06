#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

container_engine=${CONTAINER_ENGINE:-"podman"}

bazel run \
  --@rules_img//img/settings:load_daemon="${container_engine}" \
  //seed/devprod/golink/container:load

"${container_engine}" run --name "golink" --rm --interactive --tty \
  --network "host" \
  ghcr.io/ndscm/seed-devprod-golink-container:latest
