#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

tag="${1:-"$(cat ./TAG)"}"

if [[ "${1:-}" == "latest" ]]; then
  latest="$(curl -s "https://api.github.com/repos/containers/podman/releases/latest")"
  tag="$(jq -r '.tag_name | ltrimstr("v")' <<<"${latest}")"
  printf "%s\n" "${tag}" >./TAG
fi

xslash -- create --skeleton "./podman-remote.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/podman-remote.xslash" --symlink podman-remote
