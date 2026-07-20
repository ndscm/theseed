#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v5.1.0"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/lego/lego.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/lego/bin/lego
chmod +x ./seed/vendor/lego/bin/lego
