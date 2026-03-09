#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

branch=${1}
nd_user_handle=${2:-"${ND_USER_HANDLE}"}
workstation=$(hostname -s)

cd "${HOME}/theseed/${branch}"

git add --all
git commit -m "wip: devcron: changes at $(date +%Y-%dT%H:%M:%S%z)" || true
