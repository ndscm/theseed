#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

linux="${1:-"$(cat ./LINUX)"}"
version="${2:-"$(cat ./VERSION)"}"
darwin="${3:-"$(cat ./DARWIN)"}"

if [[ "${1:-}" == "latest" ]]; then
  # Official Perforce release channel: an Apache directory index of rYY.N releases.
  # The newest releases are often server-only and omit these GUI artifacts, so
  # walk the releases high-to-low and take the first that still ships each one.
  directory="https://filehost.perforce.com/perforce"

  # Release tags, highest version first.
  candidates="$(curl -fsS "${directory}/" | grep -oE 'r[0-9]+\.[0-9]+/' | tr -d '/' | sort -Vru)"

  # Linux: highest release shipping the p4v bundle.
  linux=""
  while read -r candidate; do
    if curl -fsI "${directory}/${candidate}/bin.linux26x86_64/p4v.tgz" >/dev/null 2>&1; then
      linux="${candidate}"
      break
    fi
  done <<<"${candidates}"
  [[ -n "${linux}" ]] || {
    printf "no release with bin.linux26x86_64/p4v.tgz found\n" >&2
    exit 1
  }
  version="$(
    set +o pipefail
    curl -fsSL "${directory}/${linux}/bin.linux26x86_64/p4v.tgz" 2>/dev/null |
      tar -tz 2>/dev/null |
      grep -m1 -oE '^p4v-[^/]+' |
      sed 's/^p4v-//'
  )"
  [[ -n "${version}" ]] || {
    printf "could not determine p4v version for %s\n" "${linux}" >&2
    exit 1
  }

  # macOS: highest release shipping the universal (amd64 + arm64) P4V disk image.
  darwin=""
  while read -r candidate; do
    if curl -fsI "${directory}/${candidate}/bin.macosx12u/P4V.dmg" >/dev/null 2>&1; then
      darwin="${candidate}"
      break
    fi
  done <<<"${candidates}"
  [[ -n "${darwin}" ]] || {
    printf "no release with bin.macosx12u/P4V.dmg found\n" >&2
    exit 1
  }

  printf "%s\n" "${linux}" >./LINUX
  printf "%s\n" "${version}" >./VERSION
  printf "%s\n" "${darwin}" >./DARWIN
fi

replacements=(
  --replace "{{LINUX}}=${linux}"
  --replace "{{VERSION}}=${version}"
  --replace "{{DARWIN}}=${darwin}"
)

xslash -- create --skeleton "./p4merge.xslash.json" "${replacements[@]}" --final --out "./bin/p4merge.xslash" --symlink p4merge
