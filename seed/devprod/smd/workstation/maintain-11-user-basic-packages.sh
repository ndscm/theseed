#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking basic packages...\e[0m\n"

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    mkdir -p "${HOME}/.local/bin"
    export PATH="${HOME}/.local/bin:${PATH}"

    host_kernel="$(uname -s | tr '[:upper:]' '[:lower:]')"
    host_arch="$(uname -m | tr '[:upper:]' '[:lower:]')"

    case "${host_arch}" in
    x86_64 | amd64) host_arch="amd64" ;;
    aarch64 | arm64) host_arch="arm64" ;;
    esac

    curl -fsSL "https://github.com/ndscm/xslash/releases/latest/download/xslash-${host_kernel}-${host_arch}" \
      -o "${HOME}/.local/bin/xslash"
    chmod +x "${HOME}/.local/bin/xslash"
  fi

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    curl -fsSL "https://raw.githubusercontent.com/ndscm/theseed/refs/heads/main/seed/prebuilt/direnv/bin/direnv.xslash" \
      -o "${HOME}/.local/bin/direnv.xslash"
    chmod +x "${HOME}/.local/bin/direnv.xslash"
    ln -s -f direnv.xslash "${HOME}/.local/bin/direnv"

    curl -fsSL "https://raw.githubusercontent.com/ndscm/theseed/refs/heads/main/seed/prebuilt/ndscm/bin/ndscm.xslash" \
      -o "${HOME}/.local/bin/ndscm.xslash"
    chmod +x "${HOME}/.local/bin/ndscm.xslash"
    ln -s -f ndscm.xslash "${HOME}/.local/bin/nd"
    ln -s -f ndscm.xslash "${HOME}/.local/bin/ndscm"
  fi

  printf "\e[32m[user] Check basic packages done.\e[0m\n"
fi
