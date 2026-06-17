#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

uname="$(uname)"
distro="unknown"
oslike="unknown"
wsl="false"

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

export uname
export distro
export oslike
export wsl
