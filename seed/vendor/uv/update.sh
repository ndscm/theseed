#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="0.11.21"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/uv/uv.dotslash.json" \
  --tag="${tag}" \
  >./seed/vendor/uv/bin/uv
chmod +x ./seed/vendor/uv/bin/uv

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/uv/uvx.dotslash.json" \
  --tag="${tag}" \
  >./seed/vendor/uv/bin/uvx
chmod +x ./seed/vendor/uv/bin/uvx
