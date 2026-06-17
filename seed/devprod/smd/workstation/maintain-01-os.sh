#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

uname="$(uname)"
distro="unknown"
oslike="unknown"
wsl="false"
run_sudo="false"

if [[ "${uname}" == "Linux" ]]; then
  if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    if [[ "${ID}" == "ubuntu" ]]; then
      distro="ubuntu"
      oslike="debian"
    fi
  fi
elif [[ "${uname}" == "Darwin" ]]; then
  distro="darwin"
  oslike="darwin"
fi

if [[ "${uname}" == "Linux" && "${distro}" == "ubuntu" && "${oslike}" == "debian" ]]; then
  printf "\e[33mSystem Distro: ${distro}\nPackage Manager: ${oslike}\e[0m\n"
elif [[ "${uname}" == "Darwin" && "${distro}" == "darwin" && "${oslike}" == "darwin" ]]; then
  printf "\e[33mSystem Distro: ${distro}\nPackage Manager: ${oslike}\e[0m\n"
else
  printf "\e[31mUnsupported system: ${distro} (${oslike})\e[0m\n"
  exit 1
fi

if [[ ! -z "${WSL_DISTRO_NAME+x}" ]]; then
  wsl="true"
  printf "\e[33mWSL: true\e[0m\n"
fi

if id -nG "$(id -un)" | tr ' ' '\n' | grep -qx -e sudo -e admin -e wheel; then
  read -p "You have sudo privileges. Use sudo for installation? [y/N]: " use_sudo_reply
  if [[ "${use_sudo_reply,,}" == "y" || "${use_sudo_reply,,}" == "yes" ]]; then
    run_sudo="true"
  fi
fi
printf "\e[33mUse sudo: ${run_sudo}\e[0m\n"

if [[ "${oslike}" == "darwin" ]]; then
  if ! xcode-select -p >/dev/null 2>&1; then
    printf "\e[33mXcode Command Line Tools not found. Installing...\e[0m\n"
    xcode-select --install
  fi
  if ! xcodebuild -license check >/dev/null 2>&1; then
    printf "\e[33mXcode license is NOT accepted yet...\e[0m\n"
    if [[ "${run_sudo}" == "true" ]]; then
      sudo xcodebuild -license accept
    else
      printf "\e[31mPlease accept the license with\n    sudo xcodebuild -license accept\e[0m\n"
      exit 1
    fi
  fi
fi

export uname
export distro
export oslike
export wsl
export run_sudo
