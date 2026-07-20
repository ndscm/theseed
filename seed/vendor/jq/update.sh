#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"1.8.1"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/jq/jq.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/jq/bin"

chmod +x ./seed/vendor/jq/bin/jq.dotslash

ln -s -f jq.dotslash ./seed/vendor/jq/bin/jq
