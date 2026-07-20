#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v26.7.4"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/ndscm/ndscm.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/ndscm/bin"

chmod +x ./seed/vendor/ndscm/bin/ndscm.dotslash

ln -s -f ndscm.dotslash ./seed/vendor/ndscm/bin/ndscm
