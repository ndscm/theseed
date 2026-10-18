#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"1.27.0"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/go/go.dotslash.json" \
  --skeleton "$(pwd)/seed/vendor/go/gofmt.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/go/bin"

chmod +x ./seed/vendor/go/bin/go.dotslash
chmod +x ./seed/vendor/go/bin/gofmt.dotslash

ln -s -f go.dotslash ./seed/vendor/go/bin/go
ln -s -f gofmt.dotslash ./seed/vendor/go/bin/gofmt
