#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

printf "\e[34mChecking basic packages...\e[0m\n"

if [[ "${oslike}" == "debian" ]]; then
  mkdir -p "${HOME}/.local/bin"
  if [[ "${run_sudo}" == "true" ]]; then
    sudo apt update
    sudo apt upgrade -y
    sudo apt install -y curl
    sudo apt install -y direnv
    sudo apt install -y git
    sudo apt install -y netcat-openbsd
    sudo apt install -y ssh
    sudo apt install -y tar
  else
    printf "\e[31mSkipping system package installation\e[0m\n"
  fi

  curl -fsSL "https://github.com/facebook/dotslash/releases/latest/download/dotslash-ubuntu-22.04.$(uname -m).tar.gz" | tar fxz - -C "${HOME}/.local/bin"
fi

if [[ "${oslike}" == "darwin" ]]; then
  mkdir -p "${HOME}/.local/bin"
  if [[ "${run_sudo}" == "true" ]]; then
    if brew --version; then
      printf "\e[33mFound brew, skip install homebrew.\e[0m\n"
    else
      /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    fi
    if ! grep -F -q -s -x 'eval "$(/opt/homebrew/bin/brew shellenv)"' "${HOME}/.zprofile"; then
      printf '\neval "$(/opt/homebrew/bin/brew shellenv)"\n' >>"${HOME}/.zprofile"
    fi
    eval "$(/opt/homebrew/bin/brew shellenv)"

    brew install direnv
    brew install socat
  else
    printf "\e[31mSkipping homebrew and package installation\e[0m\n"
  fi

  curl -fsSL https://github.com/facebook/dotslash/releases/latest/download/dotslash-macos.tar.gz | tar fxz - -C "${HOME}/.local/bin"
fi

printf "\e[32mCheck basic packages done.\e[0m\n"
