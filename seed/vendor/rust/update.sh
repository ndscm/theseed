#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

rustup="${1:-"1.29.1"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/rust/rustup-init.dotslash.json" \
  --replace "TAG=${rustup}" \
  --outdir "$(pwd)/seed/vendor/rust/bin"

chmod +x ./seed/vendor/rust/bin/rustup-init.dotslash

ln -s -f rustup-init.dotslash ./seed/vendor/rust/bin/rustup
ln -s -f rustup-init.dotslash ./seed/vendor/rust/bin/rustup-init
