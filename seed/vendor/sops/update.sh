#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="v3.13.2"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/sops/sops.dotslash.json" \
  --tag="${tag}" \
  --no-format \
  >./seed/vendor/sops/bin/sops
chmod +x ./seed/vendor/sops/bin/sops
