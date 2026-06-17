#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

printf "\e[34mChecking basic packages...\e[0m\n"

if [[ "${oslike}" == "debian" ]]; then
  sudo apt update
  sudo apt upgrade -y
  sudo apt install -y curl
  sudo apt install -y direnv
  sudo apt install -y git
  sudo apt install -y netcat-openbsd
  sudo apt install -y ssh
fi

if [[ "${oslike}" == "darwin" ]]; then
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
fi

printf "\e[32mCheck basic packages done.\e[0m\n"
