#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"22.1.7"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/llvm/clang-format.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/llvm/bin/clang-format
chmod +x ./seed/vendor/llvm/bin/clang-format

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/llvm/clang.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/llvm/bin/clang
chmod +x ./seed/vendor/llvm/bin/clang

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/llvm/clangd.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/llvm/bin/clangd
chmod +x ./seed/vendor/llvm/bin/clangd
