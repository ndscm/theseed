#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking bazelisk...\e[0m\n"

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    npm list --global @bazel/bazelisk || npm install --global @bazel/bazelisk
  fi

  printf "\e[32m[user] Check bazelisk done.\e[0m\n"
fi
