#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
    if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
        cat <<EOF >>"${HOME}/.managed_profile.tmp"

## Personal Profile

if [ -f "\${HOME}/.personal_profile" ]; then
  source "\${HOME}/.personal_profile"
fi
EOF
        cat <<EOF >>"${HOME}/.managed_shrc.tmp"

## Personal shrc

if [ -f "\${HOME}/.personal_shrc" ]; then
  source "\${HOME}/.personal_shrc"
fi
EOF
    fi

    mv -f "${HOME}/.managed_profile.tmp" "${HOME}/.managed_profile"
    mv -f "${HOME}/.managed_shrc.tmp" "${HOME}/.managed_shrc"

    printf "\e[32mDone. Please restart the terminal to load new environments.\e[0m\n"
fi
