#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"0.11.21"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/uv/uv.dotslash.json" \
  --skeleton "$(pwd)/seed/vendor/uv/uvx.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/uv/bin"

chmod +x ./seed/vendor/uv/bin/uv.dotslash
chmod +x ./seed/vendor/uv/bin/uvx.dotslash

ln -s -f uv.dotslash ./seed/vendor/uv/bin/uv
ln -s -f uvx.dotslash ./seed/vendor/uv/bin/uvx
