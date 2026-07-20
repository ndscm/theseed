#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"22.1.7"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/llvm/clang-format.dotslash.json" \
  --skeleton "$(pwd)/seed/vendor/llvm/clang.dotslash.json" \
  --skeleton "$(pwd)/seed/vendor/llvm/clangd.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/llvm/bin"

chmod +x ./seed/vendor/llvm/bin/clang-format.dotslash
chmod +x ./seed/vendor/llvm/bin/clang.dotslash
chmod +x ./seed/vendor/llvm/bin/clangd.dotslash

ln -s -f clang-format.dotslash ./seed/vendor/llvm/bin/clang-format
ln -s -f clang.dotslash ./seed/vendor/llvm/bin/clang
ln -s -f clangd.dotslash ./seed/vendor/llvm/bin/clangd
