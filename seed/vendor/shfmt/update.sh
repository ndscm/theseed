#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v3.13.1"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/shfmt/shfmt.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/shfmt/bin/shfmt
chmod +x ./seed/vendor/shfmt/bin/shfmt
