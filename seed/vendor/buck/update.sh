#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

buck2="${1:-"2026-05-01"}"
reindeer="${2:-"v2026.04.27.00"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/buck/buck2.dotslash.json" \
  --replace "TAG=${buck2}" \
  --outdir "$(pwd)/seed/vendor/buck/bin"

chmod +x ./seed/vendor/buck/bin/buck2.dotslash

ln -s -f buck2.dotslash ./seed/vendor/buck/bin/buck2

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/buck/reindeer.dotslash.json" \
  --replace "TAG=${reindeer}" \
  --outdir "$(pwd)/seed/vendor/buck/bin"

chmod +x ./seed/vendor/buck/bin/reindeer.dotslash

ln -s -f reindeer.dotslash ./seed/vendor/buck/bin/reindeer
