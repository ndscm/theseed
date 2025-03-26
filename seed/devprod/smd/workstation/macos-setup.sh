#!/bin/bash
set -eux
set -o pipefail

if [[ -z "${ND_USER_HANDLE+x}" ]]; then
  read -p "Enter nd user handle (username before @ndscm.com): " ND_USER_HANDLE
fi
printf "\e[33mUsername: ${ND_USER_HANDLE}\e[0m\n"
export ND_USER_HANDLE
if [[ -z "${ND_USER_DISPLAY_NAME+x}" ]]; then
  read -p "Enter user display name: " ND_USER_DISPLAY_NAME
fi
printf "\e[33mDisplay Name: ${ND_USER_DISPLAY_NAME}\e[0m\n"
export ND_USER_DISPLAY_NAME



xcode-select --install || true
sudo xcodebuild -license accept
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
if ! grep -F -x 'eval "$(/opt/homebrew/bin/brew shellenv)"' ${HOME}/.zprofile; then
  printf '\neval "$(/opt/homebrew/bin/brew shellenv)"\n' >>${HOME}/.zprofile
fi
eval "$(/opt/homebrew/bin/brew shellenv)"
brew install socat

if [[ -f "${HOME}/.ssh/id_ed25519" ]]; then
  printf "\e[33mFound .ssh/ed25519, skip regeneration.\e[0m\n"
else
  printf "\e[34mGenerating ssh key pair for ${ND_USER_HANDLE}@ndscm.com with ed25519 algorithm.\e[0m\n"
  read -p "Enter workstation tag (e.g. macmini): " workstation_tag
  ssh-keygen -t ed25519 -C "${ND_USER_HANDLE}+${workstation_tag}@ndscm.com"
  public_key=$(cat ${HOME}/.ssh/id_ed25519.pub)
  printf "\e[33mCopy your public key to your github account:\n    ${public_key}\e[0m\n"
  read -p "Press <Enter> to continue..."
fi

if [[ -z "$(sed '/Host github.com/{N;/ProxyCommand/p;}' ${HOME}/.ssh/config)" ]]; then
  printf "Found ssh proxy to github.com for git\n"
else


fi

mkdir -p ${HOME}/theseed

if [[ -d ${HOME}/theseed/theseed.git && -d ${HOME}/theseed/main && -d ${HOME}/theseed/dev ]]; then
  printf "\e[33mFound existing theseed monorepo, skip clone.\e[0m\n"
else
  if [[ -d ${HOME}/theseed/theseed.git || -d ${HOME}/theseed/main || -d ${HOME}/theseed/dev ]]; then
    printf "\e[31mFound old theseed monorepo, please backup and remove it:\e[33m
    rm -rf ${HOME}/theseed/dev
    rm -rf ${HOME}/theseed/main
    rm -rf ${HOME}/theseed/theseed.git
\e[0m"
    exit 1
  fi
  printf "\e[34mCloning theseed monorepo...\e[0m\n"
  git clone --bare --single-branch \
    --config "core.logallrefupdates=true" \
    --config "remote.origin.fetch=+refs/heads/*:refs/remotes/origin/*" \
    git@github.com:ndscm/theseed.git \
    ${HOME}/theseed/theseed.git
  git --git-dir ${HOME}/theseed/theseed.git worktree add -B main ${HOME}/theseed/main origin/main
  git --git-dir ${HOME}/theseed/theseed.git branch --track=direct base/dev origin/main
  git --git-dir ${HOME}/theseed/theseed.git branch --track=direct dev base/dev
  git --git-dir ${HOME}/theseed/theseed.git worktree add -B dev ${HOME}/theseed/dev
fi

cd ${HOME}/theseed/main
git fetch --all --prune
git rebase origin/main

if [[ -f "${HOME}/theseed/main/seed/devprod/smd/managed-workstation.sh" ]]; then
  ${HOME}/theseed/main/seed/devprod/smd/managed-workstation.sh
fi
