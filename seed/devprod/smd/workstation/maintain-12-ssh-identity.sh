#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

printf "\e[34mChecking ssh key pair...\e[0m\n"

if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
  if [[ -f "${HOME}/.ssh/id_ed25519" ]]; then
    printf "\e[33mFound .ssh/ed25519, skip regeneration.\e[0m\n"
  else
    printf "\e[34mGenerating ssh key pair for ${ND_USER_HANDLE}@${ND_USER_DOMAIN} with ed25519 algorithm.\e[0m\n"
    read -p "Enter workstation tag (e.g. t14wsl): " workstation_tag
    ssh-keygen -t ed25519 -C "${ND_USER_HANDLE}+${workstation_tag}@${ND_USER_DOMAIN}"
    public_key=$(cat "${HOME}/.ssh/id_ed25519.pub")
    printf "\e[33mCopy your public key to your github account:\n    ${public_key}\e[0m\n"
    read -p "Press <Enter> to continue..."
  fi
fi

printf "\e[32mCheck ssh key pair done.\e[0m\n"
