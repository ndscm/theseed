#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v3.13.2"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/sops/sops.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/sops/bin/sops
chmod +x ./seed/vendor/sops/bin/sops
