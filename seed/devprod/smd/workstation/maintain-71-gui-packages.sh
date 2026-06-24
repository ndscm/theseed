#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",gui,"* ]]; then
  printf "\e[34m[gui] Checking GUI packages...\e[0m\n"

  if [[ "${oslike}" == "darwin" ]]; then
    brew install --cask iterm2
    brew install --cask p4v
    brew install --cask visual-studio-code
    brew install --cask --no-quarantine chromium
  fi

  printf "\e[32m[gui] Check GUI packages done.\e[0m\n"
fi
