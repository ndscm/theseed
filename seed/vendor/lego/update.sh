#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/lego/lego.dotslash.json" \
  --tag="v5.1.0" \
  >./seed/vendor/lego/bin/lego
chmod +x ./seed/vendor/lego/bin/lego
