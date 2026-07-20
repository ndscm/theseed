#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"r26.1"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/contrib/vendor/perforce/p4merge.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/contrib/vendor/perforce/bin"

chmod +x ./contrib/vendor/perforce/bin/p4merge.dotslash

ln -s -f p4merge.dotslash ./contrib/vendor/perforce/bin/p4merge
