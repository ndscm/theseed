#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
  cat <<EOF >>"${HOME}/.seed_managed_profile.tmp"

## Personal Profile

if [ -f "\${HOME}/.personal_profile" ]; then
  source "\${HOME}/.personal_profile"
fi
EOF
  cat <<EOF >>"${HOME}/.seed_managed_shrc.tmp"

## Personal shrc

if [ -f "\${HOME}/.personal_shrc" ]; then
  source "\${HOME}/.personal_shrc"
fi
EOF
fi

mv -f "${HOME}/.seed_managed_profile.tmp" "${HOME}/.seed_managed_profile"
mv -f "${HOME}/.seed_managed_shrc.tmp" "${HOME}/.seed_managed_shrc"

printf "\e[32mDone. Please restart the terminal to load new environments.\e[0m\n"
