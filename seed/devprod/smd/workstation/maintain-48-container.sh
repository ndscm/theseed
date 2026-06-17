#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

cat <<EOF >>"${HOME}/.seed_managed_profile.tmp"

## Docker

export COMPOSE_MENU="0"
EOF
