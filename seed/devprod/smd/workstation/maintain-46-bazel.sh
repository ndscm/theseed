#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

printf "\e[34mChecking bazelisk...\e[0m\n"

if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
  npm list --global @bazel/bazelisk || npm install --global @bazel/bazelisk
fi

printf "\e[34mCheck bazelisk done.\e[0m\n"
