#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

tag="${1:-"$(cat ./TAG)"}"

if [[ "${1:-}" == "latest" ]]; then
  index="$(curl -s "https://dl.google.com/android/repository/repository2-3.xml")"
  tag="$(
    grep -oE 'platform-tools_r[0-9.]+-linux\.zip' <<<"${index}" |
      sed -E 's/^platform-tools_//; s/-linux\.zip$//' |
      sort -V |
      tail -n1
  )"
  printf "%s\n" "${tag}" >./TAG
fi

xslash -- create --skeleton "./adb.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/adb.xslash" --symlink adb
xslash -- create --skeleton "./fastboot.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/fastboot.xslash" --symlink fastboot
