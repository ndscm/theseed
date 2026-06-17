#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

printf "\e[34mChecking managed proxy...\e[0m\n"

if [[ -n "${HTTPS_PROXY:-}" && -n "${NO_PROXY:-}" ]]; then
  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    cat <<EOF >>"${HOME}/.seed_managed_profile.tmp"

## Proxy

export HTTPS_PROXY="${HTTPS_PROXY}"
export HTTP_PROXY="\${HTTPS_PROXY}"
export NO_PROXY="${NO_PROXY}"
export https_proxy="\${HTTPS_PROXY}"
export http_proxy="\${HTTP_PROXY}"
export no_proxy="\${NO_PROXY}"
EOF
  fi

  proxy_host=$(printf "%s" "${HTTPS_PROXY}" | sed -E 's#^https?://([^:/]+)(:[0-9]+)?/?#\1#')
  proxy_port=$(printf "%s" "${HTTPS_PROXY}" | sed -E 's#^https?://[^:/]+(:[0-9]+)?/?#\1#' | sed -E 's#^:([0-9]+)$#\1#')
  if [[ -z "${proxy_port}" ]]; then
    proxy_port="443"
  fi

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    if [[ -f "${HOME}/.ssh/config" && -n "$(sed -n '/Host github.com/{N;/ProxyCommand/p;}' "${HOME}/.ssh/config")" ]]; then
      printf "\e[33mFound ssh proxy to github.com for git, skip.\e[0m\n"
    else
      printf "\e[34mAdding ssh proxy to github.com for git...\e[0m\n"
      if [[ "${oslike}" == "debian" ]]; then
        printf "\n%s\n%s\n" \
          "Host github.com" \
          "    ProxyCommand nc -X connect -x ${proxy_host}:${proxy_port} ssh.github.com 443" \
          >>"${HOME}/.ssh/config"
      fi
      if [[ "${oslike}" == "darwin" ]]; then
        printf "\n%s\n%s\n" \
          "Host github.com" \
          "    ProxyCommand socat - \"PROXY:${proxy_host}:ssh.github.com:443,proxyport=${proxy_port}\"" \
          >>"${HOME}/.ssh/config"
      fi
    fi
  fi
fi

printf "\e[32mCheck managed proxy done.\e[0m\n"
