#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

container_engine=${CONTAINER_ENGINE:-"podman"}

bazel run --@rules_img//img/settings:load_daemon="${container_engine}" //seed/newtype/kurisu/container:load

"${container_engine}" run --name "kurisu" --rm --interactive --tty \
  --network "host" \
  --env SEED_OPENID_DISCOVERY_URL \
  ghcr.io/ndscm/seed-newtype-kurisu-container:latest
