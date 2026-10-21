#!/usr/bin/env bash
set -e -u -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

tag="${1:-"$(cat ./TAG)"}"
version="${2:-"$(cat ./VERSION)"}"

if [[ "${1:-}" == "latest" ]]; then
  stable="$(curl -s "https://static.rust-lang.org/dist/channel-rust-stable.toml")"
  tag="$(awk -F'"' '/^\[/{exit} /^date/{print $2; exit}' <<<"${stable}")"
  version="$(awk -F'"' '/^\[pkg\.rust\]/{f=1} f&&/^version/{split($2,v," "); print v[1]; exit}' <<<"${stable}")"
  printf "%s\n" "${tag}" >./TAG
  printf "%s\n" "${version}" >./VERSION
fi

names=(
  rustc
  rustfmt
  cargo
  cargo-clippy
  cargo-fmt
  rustdoc
  rust-gdb
)

for name in "${names[@]}"; do
  xslash -- create \
    --skeleton "./${name}.xslash.json" \
    --replace "{{TAG}}=${tag}" \
    --replace "{{VERSION}}=${version}" \
    --final true \
    --out "./bin/${name}.xslash"

  ln -s -f "${name}.xslash" "./bin/${name}"
done
