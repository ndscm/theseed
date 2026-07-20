#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v1.29.0"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/bazel/bazelisk.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/bazel/bin"

chmod +x ./seed/vendor/bazel/bin/bazelisk.dotslash

ln -s -f bazelisk.dotslash ./seed/vendor/bazel/bin/bazel
ln -s -f bazelisk.dotslash ./seed/vendor/bazel/bin/bazelisk
