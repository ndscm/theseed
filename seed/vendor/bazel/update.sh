#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v1.29.0"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/bazel/bazelisk.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/bazel/bin/bazelisk
chmod +x ./seed/vendor/bazel/bin/bazelisk
