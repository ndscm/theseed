#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",system,"* ]]; then
  printf "\e[34m[system] Checking zsh shell...\e[0m\n"

  if [[ "${oslike}" == "debian" ]]; then
    sudo -E apt install -y zsh
  fi

  printf "\e[32m[system] Check zsh shell done.\e[0m\n"
fi
