#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking node tools...\e[0m\n"

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    set +eux
    bash -c "$(curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh)"
    # Manually load NVM
    export NVM_DIR="$HOME/.nvm"
    . "$NVM_DIR/nvm.sh"
    nvm install --lts
    nvm use --lts
    set -eux
    corepack enable
  fi

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    npm list --global prettier || npm install --global prettier
    npm list --global typescript || npm install --global typescript
  fi

  printf "\e[32m[user] Check node tools done.\e[0m\n"
fi
