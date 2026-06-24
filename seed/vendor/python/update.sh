#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="3.11.15"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/python/python3.dotslash.json" \
  --tag="${tag}" \
  >./seed/vendor/python/bin/python3
chmod +x ./seed/vendor/python/bin/python3
