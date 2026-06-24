#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",system,"* ]]; then
  printf "\e[34m[system] Checking system packages...\e[0m\n"

  if [[ "${oslike}" == "debian" ]]; then
    sudo -E apt install -y build-essential
    sudo -E apt install -y clang
    sudo -E apt install -y clang-format
    sudo -E apt install -y curl
    sudo -E apt install -y davfs2
    sudo -E apt install -y default-jdk
    sudo -E apt install -y direnv
    sudo -E apt install -y g++
    sudo -E apt install -y gcc
    sudo -E apt install -y git
    sudo -E apt install -y gitg
    sudo -E apt install -y gnupg2
    sudo -E apt install -y iputils-ping
    sudo -E apt install -y jq
    sudo -E apt install -y libcap2-bin
    sudo -E apt install -y libxcb-cursor0
    sudo -E apt install -y lsb-release
    sudo -E apt install -y netcat-openbsd
    sudo -E apt install -y p7zip-full
    sudo -E apt install -y p7zip-rar
    sudo -E apt install -y python3
    sudo -E apt install -y python3-pip
    sudo -E apt install -y rsync
    sudo -E apt install -y ssh
    sudo -E apt install -y vim
    sudo -E apt install -y zsh

    sudo -E snap install --classic go
  fi

  if [[ "${oslike}" == "darwin" ]]; then
    brew install clang-format
    brew install gitg
    brew install --cask iterm2
    brew install --cask p4v
    brew install --cask visual-studio-code
    brew install --cask --no-quarantine chromium
    defaults write NSGlobalDomain ApplePressAndHoldEnabled -bool false
  fi

  printf "\e[32m[system] Check system packages done.\e[0m\n"
fi
