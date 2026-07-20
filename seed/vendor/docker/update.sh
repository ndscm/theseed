#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"29.5.3"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/docker/docker.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/docker/bin"

chmod +x ./seed/vendor/docker/bin/docker.dotslash

ln -s -f docker.dotslash ./seed/vendor/docker/bin/docker
