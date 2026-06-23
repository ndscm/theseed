#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
    printf "\e[34m[user] Checking snap tools...\e[0m\n"

    if [[ "${oslike}" == "debian" ]]; then
        cat <<EOF >>"${HOME}/.managed_profile.tmp"

## Snap

export PATH="/snap/bin:\${PATH}"
EOF
    fi

    printf "\e[32m[user] Check snap tools done.\e[0m\n"
fi
