#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

## Shell

printf "\e[34mChecking zsh shell...\e[0m\n"

if [[ "${oslike}" == "debian" ]]; then
  if [[ "${run_sudo}" == "true" ]]; then
    sudo -E apt install -y zsh
  else
    printf "\e[31mSkipping system zsh installation\e[0m\n"
  fi
fi

if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
  if [[ ! -d "${HOME}/.oh-my-zsh" ]]; then
    bash -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" -- --unattended
  fi
  cp "${HOME}/.oh-my-zsh/templates/zshrc.zsh-template" "${HOME}/.zshrc"
  sed -i.bak 's/ZSH_THEME=".*"/ZSH_THEME="agnoster"/' "${HOME}/.zshrc"
  printf '\nunsetopt SHARE_HISTORY\n' >>"${HOME}/.zshrc"
  if [[ "${SHELL}" != "/bin/zsh" ]]; then
    chsh -s /bin/zsh
  fi
fi

printf "\e[32mCheck zsh shell done.\e[0m\n"
