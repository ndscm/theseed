#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

tag="${1:-"$(cat ./TAG)"}"

if [[ "${1:-}" == "latest" ]]; then
  tag="$(curl -s "https://release.ariga.io/atlas/atlas-linux-amd64-latest.version")"
  printf "%s\n" "${tag}" >./TAG
fi

xslash -- create --skeleton "./atlas.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/atlas.xslash" --symlink atlas
