#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

tag="${1:-"$(cat ./TAG)"}"

if [[ "${1:-}" == "latest" ]]; then
  tag="$(curl -s "https://dl.k8s.io/release/stable.txt")"
  printf "%s\n" "${tag}" >./TAG
fi

xslash -- create --skeleton "./kubectl.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/kubectl.xslash" --symlink kubectl
