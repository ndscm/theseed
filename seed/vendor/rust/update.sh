#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"1.96.0"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/rust/rustup-init.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/rust/bin"

chmod +x ./seed/vendor/rust/bin/rustup-init.dotslash

ln -s -f rustup-init.dotslash ./seed/vendor/rust/bin/rustup
ln -s -f rustup-init.dotslash ./seed/vendor/rust/bin/rustup-init
