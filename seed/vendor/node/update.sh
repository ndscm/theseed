#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v24.14.1"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/node/corepack.dotslash.json" \
  --skeleton "$(pwd)/seed/vendor/node/node.dotslash.json" \
  --skeleton "$(pwd)/seed/vendor/node/npm.dotslash.json" \
  --skeleton "$(pwd)/seed/vendor/node/npx.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/node/bin"

chmod +x ./seed/vendor/node/bin/corepack.dotslash
chmod +x ./seed/vendor/node/bin/node.dotslash
chmod +x ./seed/vendor/node/bin/npm.dotslash
chmod +x ./seed/vendor/node/bin/npx.dotslash

ln -s -f corepack.dotslash ./seed/vendor/node/bin/corepack
ln -s -f node.dotslash ./seed/vendor/node/bin/node
ln -s -f npm.dotslash ./seed/vendor/node/bin/npm
ln -s -f npx.dotslash ./seed/vendor/node/bin/npx
ln -s -f pnpm.shim ./seed/vendor/node/bin/pnpm
ln -s -f pnpx.shim ./seed/vendor/node/bin/pnpx
ln -s -f yarn.shim ./seed/vendor/node/bin/yarn
ln -s -f yarnpkg.shim ./seed/vendor/node/bin/yarnpkg
