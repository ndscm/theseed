#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  cat <<EOF >>"${HOME}/.managed_profile.tmp"

## Docker

export COMPOSE_MENU="0"
EOF
fi
