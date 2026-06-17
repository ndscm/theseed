#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

printf "\e[34mChecking snap tools...\e[0m\n"

if [[ "${oslike}" == "debian" ]]; then
  cat <<EOF >>"${HOME}/.seed_managed_profile.tmp"

## Snap

export PATH="/snap/bin:\${PATH}"
EOF
fi

printf "\e[32mCheck snap tools done.\e[0m\n"
