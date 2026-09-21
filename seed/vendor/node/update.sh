#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v26.7.0"}"
pnpm="${2:-"v12.5.1"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/node/node.dotslash.json" \
  --skeleton "$(pwd)/seed/vendor/node/npm.dotslash.json" \
  --skeleton "$(pwd)/seed/vendor/node/npx.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/node/bin"

chmod +x ./seed/vendor/node/bin/node.dotslash
chmod +x ./seed/vendor/node/bin/npm.dotslash
chmod +x ./seed/vendor/node/bin/npx.dotslash

ln -s -f node.dotslash ./seed/vendor/node/bin/node
ln -s -f npm.dotslash ./seed/vendor/node/bin/npm
ln -s -f npx.dotslash ./seed/vendor/node/bin/npx

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/node/pnpm.dotslash.json" \
  --replace "TAG=${pnpm}" \
  --outdir "$(pwd)/seed/vendor/node/bin"

chmod +x ./seed/vendor/node/bin/pnpm.dotslash

ln -s -f pnpm.dotslash ./seed/vendor/node/bin/pnpm
