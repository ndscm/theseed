#!/usr/bin/env bash
set -euxo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

tag="${1:-"$(cat ./TAG)"}"
version="${2:-"$(cat ./VERSION)"}"

if [[ "${1:-}" == "latest" ]]; then
  latest="$(curl -s "https://api.github.com/repos/astral-sh/python-build-standalone/releases/latest")"
  tag="$(jq -r ".tag_name" <<<"${latest}")"
  asset_names="$(jq -r ".assets[].name" <<<"${latest}")"
  version="$(
    awk 'match($0, /cpython-[0-9]+\.[0-9]+\.[0-9]+[+]/) {
      v = substr($0, RSTART, RLENGTH)
      gsub(/^cpython-|[+]$/, "", v)
      print v
    }' <<<"${asset_names}" |
      sort -t. -k1,1n -k2,2n -k3,3n -u |
      tail -n1
  )"
  printf "%s\n" "${tag}" >./TAG
  printf "%s\n" "${version}" >./VERSION
elif [[ "${1:-}" == "lts" ]]; then
  latest="$(curl -s "https://api.github.com/repos/astral-sh/python-build-standalone/releases/latest")"
  tag="$(jq -r ".tag_name" <<<"${latest}")"
  asset_names="$(jq -r ".assets[].name" <<<"${latest}")"
  minor="$(
    curl -s "https://peps.python.org/api/release-cycle.json" |
      jq -r 'to_entries[] | select(.value.status == "security") | .key' |
      sort -t. -k1,1n -k2,2n |
      tail -n1
  )"
  version="$(
    awk -v prefix="cpython-${minor}." 'index($0, prefix) && match($0, /cpython-[0-9]+\.[0-9]+\.[0-9]+[+]/) {
      v = substr($0, RSTART, RLENGTH)
      if (index(v, prefix) != 1) next
      gsub(/^cpython-|[+]$/, "", v)
      print v
    }' <<<"${asset_names}" |
      sort -t. -k1,1n -k2,2n -k3,3n -u |
      tail -n1
  )"
  printf "%s\n" "${tag}" >./TAG
  printf "%s\n" "${version}" >./VERSION
fi

replacements=(--replace "{{TAG}}=${tag}" --replace "{{VERSION}}=${version}")

xslash -- create --skeleton "./python3.xslash.json" "${replacements[@]}" --final --out "./bin/python3.xslash" --symlink python --symlink python3
