#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v0.21.6"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/crane/crane.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/crane/bin"

chmod +x ./seed/vendor/crane/bin/crane.dotslash

ln -s -f crane.dotslash ./seed/vendor/crane/bin/crane
