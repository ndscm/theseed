#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  all_proxy="${1:-}"
  extra_no_proxy="${2:-}"

  printf "\e[34mChecking proxy configs...\e[0m\n"

  if [[ -z "${all_proxy}" ]]; then
    printf "\e[33mNo proxy target specified.\e[0m\n"
    exit 1
  fi

  export HTTPS_PROXY="${all_proxy}"
  export HTTP_PROXY="${HTTPS_PROXY}"
  export NO_PROXY="127.0.0.1,127.0.0.0/8,192.168.0.0/16,172.16.0.0/12,10.0.0.0/8,localhost"

  if [[ -n "${extra_no_proxy}" ]]; then
    export NO_PROXY="${NO_PROXY},${extra_no_proxy}"
  fi

  export https_proxy="${HTTPS_PROXY}"
  export http_proxy="${HTTP_PROXY}"
  export no_proxy="${NO_PROXY}"

  printf "\e[32mCheck proxy configs done.\e[0m\n"
fi
