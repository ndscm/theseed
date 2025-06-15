#!/usr/bin/env bash
set -eux
set -o pipefail

# # Detect distro

uname="$(uname)"
distro="unknown"
oslike="unknown"

if [[ "${uname}" == "Darwin" ]]; then
  distro="darwin"
  oslike="darwin"
fi
if [[ "${uname}" == "Linux" ]]; then
  if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    if [[ "${ID}" == "debian" ]]; then
      distro="debian"
      oslike="debian"
    fi
    if [[ "${ID}" == "fedora" ]]; then
      distro="fedora"
      oslike="fedora"
    fi
    if [[ "${ID}" == "rocky" ]]; then
      distro="rocky"
      oslike="fedora"
    fi
    if [[ "${ID}" == "ubuntu" ]]; then
      distro="ubuntu"
      oslike="debian"
    fi
  fi
fi

if [[ "${uname}" == "Linux" && "${distro}" == "fedora" && "${oslike}" == "fedora" ]]; then
  : # support fedora
elif [[ "${uname}" == "Linux" && "${distro}" == "rocky" && "${oslike}" == "fedora" ]]; then
  : # support rocky
else
  printf "\e[31mUnsupported system: ${distro} (${oslike})\e[0m\n"
  exit 1
fi
printf "\e[33mSystem Distro: ${distro}\nPackage Manager: ${oslike}\e[0m\n"

# # Create theseed user and group

if [[ "${oslike}" == "fedora" ]]; then
  useradd --create-home theseed
  usermod -a -G wheel theseed
fi
