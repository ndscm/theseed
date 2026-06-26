#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking zsh shell...\e[0m\n"

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    if [[ ! -d "${HOME}/.oh-my-zsh" ]]; then
      bash -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" -- --unattended
    fi
    cp "${HOME}/.oh-my-zsh/templates/zshrc.zsh-template" "${HOME}/.zshrc"
    sed -i.bak 's/ZSH_THEME=".*"/ZSH_THEME="agnoster"/' "${HOME}/.zshrc"
    printf '\nunsetopt SHARE_HISTORY\n' >>"${HOME}/.zshrc"
  fi

  login_shell=""
  if [[ "${oslike}" == "debian" ]]; then
    login_shell="$(getent passwd "$(id -un)" | cut -d: -f7)"
  fi
  if [[ "${oslike}" == "darwin" ]]; then
    login_shell="$(dscl . -read "/Users/$(id -un)" UserShell | awk '{print $2}')"
  fi

  if [[ "${login_shell}" != */zsh ]]; then
    read -r -p "Change login shell to $(which zsh)? [y/N] " reply
    if [[ "${reply}" == [yY] ]]; then
      chsh -s "$(which zsh)"
    fi
  fi

  printf "\e[32m[user] Check zsh shell done.\e[0m\n"
fi
