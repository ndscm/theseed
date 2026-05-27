#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/bazel/bazelisk.dotslash.json" \
  --no-format \
  --tag="v1.29.0" \
  >./seed/vendor/bazel/bin/bazelisk
chmod +x ./seed/vendor/bazel/bin/bazelisk
