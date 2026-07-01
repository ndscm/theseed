#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="v2.37.1"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/direnv/direnv.dotslash.json" \
  --no-format \
  --tag="${tag}" \
  >./seed/vendor/direnv/bin/direnv
chmod +x ./seed/vendor/direnv/bin/direnv
