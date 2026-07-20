#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v1.3.1"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/age/age.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/age/bin/age
chmod +x ./seed/vendor/age/bin/age

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/age/age-keygen.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/age/bin/age-keygen
chmod +x ./seed/vendor/age/bin/age-keygen

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/age/age-inspect.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/age/bin/age-inspect
chmod +x ./seed/vendor/age/bin/age-inspect
