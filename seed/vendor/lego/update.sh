#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v5.1.0"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/lego/lego.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/lego/bin"

chmod +x ./seed/vendor/lego/bin/lego.dotslash

ln -s -f lego.dotslash ./seed/vendor/lego/bin/lego
