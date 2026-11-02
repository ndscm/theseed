#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

tag="${1:-"$(cat ./TAG)"}"

if [[ "${1:-}" == "latest" ]]; then
  index="$(curl -s "https://download.docker.com/linux/static/stable/x86_64/")"
  tag="$(
    grep -oE 'docker-[0-9]+\.[0-9]+\.[0-9]+\.tgz' <<<"${index}" |
      sed -E 's/^docker-//; s/\.tgz$//' |
      sort -t. -k1,1n -k2,2n -k3,3n -u |
      tail -n1
  )"
  printf "%s\n" "${tag}" >./TAG
fi

xslash -- create --skeleton "./docker.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/docker.xslash" --symlink docker
