#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

tag="${1:-"$(cat ./TAG)"}"

if [[ "${1:-}" == "latest" ]]; then
  tag="$(curl -fsSL "https://downloads.claude.ai/claude-code-releases/latest")"
  printf "%s\n" "${tag}" >./TAG
fi

xslash -- create --skeleton "./claude.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/claude.xslash" --symlink claude
