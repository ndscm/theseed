#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="v3.13.1"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/shfmt/shfmt.dotslash.json" \
  --tag="${tag}" \
  --no-format \
  >./seed/vendor/shfmt/bin/shfmt
chmod +x ./seed/vendor/shfmt/bin/shfmt
