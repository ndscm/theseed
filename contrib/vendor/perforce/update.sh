#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="r26.1"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/contrib/vendor/perforce/p4merge.dotslash.json" \
  --tag="${tag}" \
  >./contrib/vendor/perforce/bin/p4merge
chmod +x ./contrib/vendor/perforce/bin/p4merge
