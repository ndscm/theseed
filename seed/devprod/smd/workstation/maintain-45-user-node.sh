#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking node tools...\e[0m\n"

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    # Install global packages under ~/npm, mirroring how Go handles global packages
    npm config set prefix "~/npm"
    cat <<EOF >>"${HOME}/.managed_profile.tmp"

## NPM

export PATH="\${HOME}/npm/bin:\${PATH}"
EOF

    npm list --global prettier || npm install --global prettier
    npm list --global typescript || npm install --global typescript
  fi

  printf "\e[32m[user] Check node tools done.\e[0m\n"
fi
