#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"1.8.1"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/jq/jq.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/jq/bin/jq
chmod +x ./seed/vendor/jq/bin/jq
