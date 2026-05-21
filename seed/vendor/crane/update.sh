#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/crane/crane.dotslash.json" \
  --tag="v0.21.6" \
  >./seed/vendor/crane/bin/crane
chmod +x ./seed/vendor/crane/bin/crane
