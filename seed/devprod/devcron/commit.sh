#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

monorepo_home="${1}"
user_handle="${2}"
focus="${3}"

cd "${monorepo_home}/${user_handle}/dev/${focus}"

git add --all
git commit -m "wip: devcron: changes at $(date +%Y-%m-%dT%H:%M:%S%z)" || true
