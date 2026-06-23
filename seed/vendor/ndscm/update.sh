#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="v26.6.23"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/ndscm/ndscm.dotslash.json" \
  --tag="${tag}" \
  >./seed/vendor/ndscm/bin/ndscm
chmod +x ./seed/vendor/ndscm/bin/ndscm
