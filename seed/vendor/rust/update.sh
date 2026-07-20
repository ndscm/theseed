#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"1.96.0"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/rust/rustup-init.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/rust/bin/rustup-init
chmod +x ./seed/vendor/rust/bin/rustup-init
