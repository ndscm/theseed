#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"v1.3.1"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/age/age-inspect.dotslash.json" \
  --skeleton "$(pwd)/seed/vendor/age/age-keygen.dotslash.json" \
  --skeleton "$(pwd)/seed/vendor/age/age.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/age/bin"

chmod +x ./seed/vendor/age/bin/age-inspect.dotslash
chmod +x ./seed/vendor/age/bin/age-keygen.dotslash
chmod +x ./seed/vendor/age/bin/age.dotslash

ln -s -f age-inspect.dotslash ./seed/vendor/age/bin/age-inspect
ln -s -f age-keygen.dotslash ./seed/vendor/age/bin/age-keygen
ln -s -f age.dotslash ./seed/vendor/age/bin/age
