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

export uname
export distro
export oslike
export wsl

maintain_scopes="${maintain_scopes:-""}"

if [[ "${maintain_scopes}" == "" ]]; then
  maintain_scopes="user"
  if id -nG "$(id -un)" | tr ' ' '\n' | grep -qx -e sudo -e admin -e wheel; then
    read -p "You have sudo privileges. Use sudo for system maintenance? [y/N]: " use_sudo_reply
    if [[ "${use_sudo_reply,,}" == "y" || "${use_sudo_reply,,}" == "yes" ]]; then
      maintain_scopes="system,user"
    fi
  fi
fi

export maintain_scopes
