#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v26.7.4"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/ndscm/ndscm.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/ndscm/bin/ndscm
chmod +x ./seed/vendor/ndscm/bin/ndscm
