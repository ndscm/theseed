#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="v1.3.1"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/age/age.dotslash.json" \
  --tag="${tag}" \
  >./seed/vendor/age/bin/age
chmod +x ./seed/vendor/age/bin/age

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/age/age-keygen.dotslash.json" \
  --tag="${tag}" \
  >./seed/vendor/age/bin/age-keygen
chmod +x ./seed/vendor/age/bin/age-keygen

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/age/age-inspect.dotslash.json" \
  --tag="${tag}" \
  >./seed/vendor/age/bin/age-inspect
chmod +x ./seed/vendor/age/bin/age-inspect
