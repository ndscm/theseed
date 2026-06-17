#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

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

sudo apt update
sudo apt upgrade -y
sudo apt install -y git
sudo apt install -y netcat-openbsd
sudo apt install -y ssh

if [[ -f "${HOME}/.ssh/id_ed25519" ]]; then
  echo -e "\e[33mFound .ssh/ed25519, skip regeneration.\e[0m"
else
  echo -e "\e[34mGenerating ssh key pair for ${ND_USER_HANDLE}@ndscm.com with ed25519 algorithm.\e[0m"
  read -p "Enter workstation tag (e.g. t14wsl): " workstation_tag
  ssh-keygen -t ed25519 -C "${ND_USER_HANDLE}+${workstation_tag}@ndscm.com"
  public_key=$(cat ${HOME}/.ssh/id_ed25519.pub)
  echo -e "\e[33mCopy your public key to your github account:\n    ${public_key}\e[0m"
  read -p "Press <Enter> to continue..."
fi

mkdir -p ${HOME}/theseed

if [[ -d ${HOME}/theseed/theseed.git && -d ${HOME}/theseed/main && -d ${HOME}/theseed/dev ]]; then
  echo -e "\e[33mFound existing theseed monorepo, skip clone.\e[0m"
else
  echo -e "\e[34mCloning theseed monorepo...\e[0m"
  git clone --bare --single-branch \
    --config "core.logallrefupdates=true" \
    --config "remote.origin.fetch=+refs/heads/*:refs/remotes/origin/*" \
    git@github.com:ndscm/theseed.git \
    ${HOME}/theseed/theseed.git
  git --git-dir ${HOME}/theseed/theseed.git worktree add -B main ${HOME}/theseed/main origin/main
fi

cd ${HOME}/theseed/main
git fetch --all --prune
git rebase origin/main

if [[ -f "${HOME}/theseed/main/seed/devprod/smd/managed-workstation.sh" ]]; then
  ${HOME}/theseed/main/seed/devprod/smd/managed-workstation.sh
fi
