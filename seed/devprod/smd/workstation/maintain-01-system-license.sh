#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",system,"* ]]; then
  if [[ "${oslike}" == "darwin" ]]; then
    if ! xcode-select -p >/dev/null 2>&1; then
      printf "\e[33mXcode Command Line Tools not found. Installing...\e[0m\n"
      xcode-select --install
    fi
    if ! xcodebuild -license check >/dev/null 2>&1; then
      printf "\e[33mXcode license is NOT accepted yet...\e[0m\n"
      sudo xcodebuild -license accept
    fi
  fi
fi
