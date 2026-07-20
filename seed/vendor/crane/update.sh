#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v0.21.6"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/crane/crane.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/crane/bin/crane
chmod +x ./seed/vendor/crane/bin/crane
