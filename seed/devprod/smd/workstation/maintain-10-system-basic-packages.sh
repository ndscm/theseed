#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",system,"* ]]; then
  printf "\e[34m[system] Checking basic packages...\e[0m\n"

  if [[ "${oslike}" == "debian" ]]; then
    sudo -E apt update
    sudo -E apt upgrade -y
    sudo -E apt install -y curl
    sudo -E apt install -y direnv
    sudo -E apt install -y git
    sudo -E apt install -y netcat-openbsd
    sudo -E apt install -y ssh
    sudo -E apt install -y tar
  fi

  if [[ "${oslike}" == "darwin" ]]; then
    if brew --version; then
      printf "\e[33mFound brew, skip install homebrew.\e[0m\n"
    else
      /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    fi
    if ! grep -F -q -s -x 'eval "$(/opt/homebrew/bin/brew shellenv)"' "${HOME}/.zprofile"; then
      printf '\neval "$(/opt/homebrew/bin/brew shellenv)"\n' >>"${HOME}/.zprofile"
    fi
    eval "$(/opt/homebrew/bin/brew shellenv)"

    brew install direnv
    brew install socat

    # `brew shellenv` prepends /opt/homebrew/bin to PATH, shadowing the repo's
    # dotslash-managed toolchains (e.g. go) for the rest of this run. Reload
    # direnv so the repo-managed PATH entries take precedence again.
    if command -v direnv >/dev/null 2>&1; then
      direnv reload >/dev/null 2>&1 || true
      eval "$(direnv export bash)"
    fi
  fi

  printf "\e[32m[system] Check basic packages done.\e[0m\n"
fi
