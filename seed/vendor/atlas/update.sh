#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v1.3.0"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/atlas/atlas.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/atlas/bin"

chmod +x ./seed/vendor/atlas/bin/atlas.dotslash

ln -s -f atlas.dotslash ./seed/vendor/atlas/bin/atlas
