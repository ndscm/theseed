#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="1.26.4"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/go/go.dotslash.json" \
  --tag="${tag}" \
  >./seed/vendor/go/bin/go
chmod +x ./seed/vendor/go/bin/go
