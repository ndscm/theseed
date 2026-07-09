#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking theseed monorepo...\e[0m\n"

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    mkdir -p "${HOME}/ndscm"

    if [[ -d "${HOME}/ndscm/theseed" ]]; then
      printf "\e[33mFound existing theseed monorepo, skip connect.\e[0m\n"
    else
      ndscm connect theseed "git@github.com:ndscm/theseed.git"
    fi
  fi

  printf "\e[32m[user] Check theseed monorepo done.\e[0m\n"
fi
