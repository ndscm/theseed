#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"20260623"}"
version="${2:-"3.11.15"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/python/python3.dotslash.json" \
  --replace "TAG=${tag}" \
  --replace "VERSION=${version}" \
  --outdir "$(pwd)/seed/vendor/python/bin"

chmod +x ./seed/vendor/python/bin/python3.dotslash

ln -s -f python3.dotslash ./seed/vendor/python/bin/python
ln -s -f python3.dotslash ./seed/vendor/python/bin/python3
