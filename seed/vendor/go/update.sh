#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"1.26.5"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/go/go.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/go/bin"

chmod +x ./seed/vendor/go/bin/go.dotslash

ln -s -f go.dotslash ./seed/vendor/go/bin/go
