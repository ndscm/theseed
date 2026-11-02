#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

tag="${1:-"$(cat ./TAG)"}"

if [[ "${1:-}" == "latest" ]]; then
  latest="$(curl -s "https://api.github.com/repos/llvm/llvm-project/releases/latest")"
  tag="$(jq -r '.tag_name | ltrimstr("llvmorg-")' <<<"${latest}")"
  printf "%s\n" "${tag}" >./TAG
fi

xslash -- create --skeleton "./clang.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/clang.xslash" --symlink clang
xslash -- create --skeleton "./clang-format.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/clang-format.xslash" --symlink clang-format
xslash -- create --skeleton "./clangd.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/clangd.xslash" --symlink clangd
