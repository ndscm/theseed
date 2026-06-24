#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking basic packages...\e[0m\n"

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    mkdir -p "${HOME}/.local/bin"
    export PATH="${HOME}/.local/bin:${PATH}"
    cp -f ./seed/vendor/ndscm/bin/ndscm "${HOME}/.local/bin/ndscm"
  fi

  if [[ "${oslike}" == "debian" ]]; then
    curl -fsSL "https://github.com/facebook/dotslash/releases/latest/download/dotslash-ubuntu-22.04.$(uname -m).tar.gz" |
      tar fxz - -C "${HOME}/.local/bin"
  fi

  if [[ "${oslike}" == "darwin" ]]; then
    curl -fsSL https://github.com/facebook/dotslash/releases/latest/download/dotslash-macos.tar.gz |
      tar fxz - -C "${HOME}/.local/bin"
  fi

  printf "\e[32m[user] Check basic packages done.\e[0m\n"
fi
