#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

tag="${1:-"$(cat ./TAG)"}"

if [[ "${1:-}" == "latest" ]]; then
  stable="$(curl -s "https://static.rust-lang.org/rustup/release-stable.toml")"
  tag="$(awk -F"'" '/^version/{print $2; exit}' <<<"${stable}")"
  printf "%s\n" "${tag}" >./TAG
fi

xslash -- create --skeleton "./rustup-init.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/rustup-init.xslash" --symlink rustup --symlink rustup-init
