#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

# Tag is the version without the leading "v" so {{TAG}} also matches the
# version embedded in the asset names and the dir inside the archives.
tag="2.94.0"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/github/gh.dotslash.json" \
  --tag="${tag}" \
  >./seed/vendor/github/bin/gh
chmod +x ./seed/vendor/github/bin/gh
