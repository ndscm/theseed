#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

printf "\e[34mChecking python tools...\e[0m\n"

if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
  curl -LsSf https://astral.sh/uv/install.sh | sh
fi

printf "\e[32mCheck python tools done.\e[0m\n"
