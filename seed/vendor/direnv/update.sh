#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v2.37.1"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/direnv/direnv.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/direnv/bin"

chmod +x ./seed/vendor/direnv/bin/direnv.dotslash

ln -s -f direnv.dotslash ./seed/vendor/direnv/bin/direnv
