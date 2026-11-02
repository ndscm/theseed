#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

tag="${1:-"$(cat ./TAG)"}"

if [[ "${1:-}" == "latest" ]]; then
  latest="$(curl -s "https://dl.google.com/dl/cloudsdk/channels/rapid/components-2.json")"
  tag="$(jq -r ".version" <<<"${latest}")"
  printf "%s\n" "${tag}" >./TAG
fi

xslash -- create --skeleton "./gcloud.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/gcloud.xslash" --symlink gcloud
