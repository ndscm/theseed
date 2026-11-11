#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

tag="${1:-"$(cat ./TAG)"}"

if [[ "${1:-}" == "latest" ]]; then
  latest="$(curl -s "https://api.github.com/repos/ip7z/7zip/releases/latest")"
  tag="$(jq -r ".tag_name" <<<"${latest}")"
  printf "%s\n" "${tag}" >./TAG
fi

suffix="${tag//./}"

replacements=(--replace "{{TAG}}=${tag}" --replace "{{SUFFIX}}=${suffix}")

xslash -- create --skeleton "./7zz.xslash.json" "${replacements[@]}" --final --out "./bin/7zz.xslash" --symlink 7z --symlink 7zz
