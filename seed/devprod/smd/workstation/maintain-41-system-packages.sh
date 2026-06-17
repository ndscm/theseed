#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

printf "\e[34mChecking system packages...\e[0m\n"

if [[ "${oslike}" == "debian" ]]; then
  if [[ "${run_sudo}" == "true" ]]; then
    sudo apt install -y build-essential
    sudo apt install -y clang
    sudo apt install -y clang-format
    sudo apt install -y curl
    sudo apt install -y davfs2
    sudo apt install -y default-jdk
    sudo apt install -y direnv
    sudo apt install -y g++
    sudo apt install -y gcc
    sudo apt install -y git
    sudo apt install -y gitg
    sudo apt install -y gnupg2
    sudo apt install -y iputils-ping
    sudo apt install -y jq
    sudo apt install -y libcap2-bin
    sudo apt install -y libxcb-cursor0
    sudo apt install -y lsb-release
    sudo apt install -y netcat-openbsd
    sudo apt install -y p7zip-full
    sudo apt install -y p7zip-rar
    sudo apt install -y python3
    sudo apt install -y python3-pip
    sudo apt install -y rsync
    sudo apt install -y ssh
    sudo apt install -y vim
    sudo apt install -y zsh
  else
    printf "\e[31mSkipping system package installation\e[0m\n"
  fi
elif [[ "${oslike}" == "darwin" ]]; then
  if [[ "${run_sudo}" == "true" ]]; then
    brew install clang-format
    brew install gitg
    brew install go
    brew install --cask iterm2
    brew install --cask p4v
    brew install --cask visual-studio-code
    brew install --cask --no-quarantine chromium
    defaults write NSGlobalDomain ApplePressAndHoldEnabled -bool false
  else
    printf "\e[31mSkipping system package installation\e[0m\n"
  fi
fi

printf "\e[32mCheck system packages done.\e[0m\n"
