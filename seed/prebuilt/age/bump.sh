#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

tag="${1:-"$(cat ./TAG)"}"

if [[ "${1:-}" == "latest" ]]; then
  latest="$(curl -s "https://api.github.com/repos/FiloSottile/age/releases/latest")"
  tag="$(jq -r ".tag_name" <<<"${latest}")"
  printf "%s\n" "${tag}" >./TAG
fi

xslash -- create --skeleton "./age.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/age.xslash" --symlink age
xslash -- create --skeleton "./age-keygen.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/age-keygen.xslash" --symlink age-keygen
xslash -- create --skeleton "./age-inspect.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/age-inspect.xslash" --symlink age-inspect
