#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v24.14.1"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/node/node.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/node/bin/node
chmod +x ./seed/vendor/node/bin/node

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/node/npm.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/node/bin/npm
chmod +x ./seed/vendor/node/bin/npm

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/node/npx.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/node/bin/npx
chmod +x ./seed/vendor/node/bin/npx

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/node/corepack.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/node/bin/corepack
chmod +x ./seed/vendor/node/bin/corepack
