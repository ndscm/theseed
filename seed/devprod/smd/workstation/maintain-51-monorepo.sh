#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

printf "\e[34mChecking theseed monorepo...\e[0m\n"

if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
  mkdir -p "${HOME}/theseed"

  if [[ -d "${HOME}/theseed/theseed.git" && -d "${HOME}/theseed/main" ]]; then
    printf "\e[33mFound existing theseed monorepo, skip clone.\e[0m\n"
  else
    printf "\e[34mCloning theseed monorepo...\e[0m\n"
    git clone --bare --single-branch \
      --config "core.logallrefupdates=true" \
      --config "remote.origin.fetch=+refs/heads/*:refs/remotes/origin/*" \
      git@github.com:ndscm/theseed.git \
      "${HOME}/theseed/theseed.git"
    git --git-dir "${HOME}/theseed/theseed.git" worktree add -B main "${HOME}/theseed/main" origin/main
  fi
  git --git-dir "${HOME}/theseed/theseed.git" config user.name "${ND_USER_DISPLAY_NAME}"
  git --git-dir "${HOME}/theseed/theseed.git" config user.email "${ND_USER_HANDLE}@${ND_USER_DOMAIN}"

  cat <<EOF >>"${HOME}/.seed_managed_profile.tmp"

## Monorepo

export ND_MONOREPO_HOME="\${HOME}/theseed"
export ND_MONOREPO_GIT_DIR="\$ND_MONOREPO_HOME/theseed.git"
EOF

  git -C "${HOME}/theseed/main" fetch --all --prune
  git -C "${HOME}/theseed/main" rebase origin/main
fi

export ND_MONOREPO_HOME="${HOME}/theseed"
export ND_MONOREPO_GIT_DIR="${ND_MONOREPO_HOME}/theseed.git"

printf "\e[34mCheck theseed monorepo done.\e[0m\n"
