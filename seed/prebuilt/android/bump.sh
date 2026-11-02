#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

tag="${1:-"$(cat ./TAG)"}"

if [[ "${1:-}" == "latest" ]]; then
  index="$(curl -s "https://dl.google.com/android/repository/repository2-3.xml")"
  tag="$(
    grep -oE 'commandlinetools-linux-[0-9]+_latest\.zip' <<<"${index}" |
      grep -oE '[0-9]+' |
      sort -n |
      tail -n1
  )"
  printf "%s\n" "${tag}" >./TAG
fi

xslash -- create --skeleton "./android.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/android.xslash" --symlink android
xslash -- create --skeleton "./apkanalyzer.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/apkanalyzer.xslash" --symlink apkanalyzer
xslash -- create --skeleton "./avdmanager.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/avdmanager.xslash" --symlink avdmanager
xslash -- create --skeleton "./sdkmanager.xslash.json" --replace "{{TAG}}=${tag}" --final --out "./bin/sdkmanager.xslash" --symlink sdkmanager
