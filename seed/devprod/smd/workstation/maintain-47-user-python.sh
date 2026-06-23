#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking python tools...\e[0m\n"

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    curl -LsSf https://astral.sh/uv/install.sh | sh
  fi

  printf "\e[32m[user] Check python tools done.\e[0m\n"
fi
