#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

printf "\e[34mChecking powerline fonts...\e[0m\n"

if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
  if [[ "${wsl}" != "true" && ! -d "${HOME}/github/powerline" ]]; then
    mkdir -p "${HOME}/github/powerline"
    curl -o "${HOME}/github/powerline/fonts.tar.gz" -L https://github.com/powerline/fonts/archive/refs/heads/master.tar.gz
    if [[ -d "${HOME}/github/powerline/fonts" ]]; then
      rm -r "${HOME}/github/powerline/fonts"
    fi
    mkdir -p "${HOME}/github/powerline/fonts"
    tar -z -x -v --strip-components 1 -f "${HOME}/github/powerline/fonts.tar.gz" -C "${HOME}/github/powerline/fonts/"
    "${HOME}/github/powerline/fonts/install.sh"
  fi
fi

printf "\e[32mCheck powerline fonts done.\e[0m\n"
