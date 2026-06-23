#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  if [[ -z "${ND_USER_HANDLE+x}" ]]; then
    read -p "Enter user handle (username before @company.com): " ND_USER_HANDLE
  fi
  printf "\e[33mUser handle: ${ND_USER_HANDLE}\e[0m\n"

  if [[ -z "${ND_USER_DOMAIN+x}" ]]; then
    read -p "Enter user domain (domain after ${ND_USER_HANDLE}@): " ND_USER_DOMAIN
  fi
  printf "\e[33mUser email: ${ND_USER_HANDLE}@${ND_USER_DOMAIN}\e[0m\n"

  if [[ -z "${ND_USER_DISPLAY_NAME+x}" ]]; then
    read -p "Enter user display name: " ND_USER_DISPLAY_NAME
  fi
  printf "\e[33mUser display Name: ${ND_USER_DISPLAY_NAME}\e[0m\n"

  export ND_USER_HANDLE
  export ND_USER_DOMAIN
  export ND_USER_EMAIL="${ND_USER_HANDLE}@${ND_USER_DOMAIN}"
  export ND_USER_DISPLAY_NAME
fi
