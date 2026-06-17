#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="29.5.3"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/docker/docker.dotslash.json" \
  --tag="${tag}" \
  >./seed/vendor/docker/bin/docker
chmod +x ./seed/vendor/docker/bin/docker
