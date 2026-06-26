#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking rust tools...\e[0m\n"

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    cat <<EOF >>"${HOME}/.managed_profile.tmp"

## Rust

if [[ -f "\${HOME}/.cargo/env" ]]; then
  . "\${HOME}/.cargo/env"
fi
EOF
    rustup default stable
    rustup update stable
  fi

  printf "\e[32m[user] Check rust tools done.\e[0m\n"
fi
