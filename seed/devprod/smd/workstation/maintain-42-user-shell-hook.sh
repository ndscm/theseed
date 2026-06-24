#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
    printf "\e[34m[user] Checking shell hooks...\e[0m\n"

    if [[ "${oslike}" == "debian" ]]; then
        cat <<EOF >>"${HOME}/.managed_profile.tmp"

## Snap

case ":\${PATH}:" in
*:/snap/bin:*) ;;
*) export PATH="/snap/bin:\${PATH}" ;;
esac
EOF
    fi

    if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
        cat <<EOF >>"${HOME}/.managed_shrc.tmp"

## Direnv

if command -v direnv >/dev/null 2>&1; then
  if [ -n "\${ZSH_VERSION}" ]; then
    eval "\$(direnv hook zsh)"
  elif [ -n "\${BASH_VERSION}" ]; then
    eval "\$(direnv hook bash)"
  fi
fi

## Ndscm

if command -v ndscm >/dev/null 2>&1; then
  eval "\$(ndscm --shell-eval shell)"
fi
EOF
    fi

    printf "\e[32m[user] Check shell hooks done.\e[0m\n"
fi
