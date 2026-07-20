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
  >./seed/vendor/python/bin/python3
chmod +x ./seed/vendor/python/bin/python3
