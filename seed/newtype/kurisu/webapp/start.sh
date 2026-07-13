#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

stack=${1:-"local"}

export HOOIN_DICTATE_SERVICE_SERVER="http://127.0.0.1:5874"
export HOOIN_INVADE_SERVICE_SERVER="http://127.0.0.1:5874"
export HOOIN_RAID_SERVICE_SERVER="http://127.0.0.1:5874"
export HOOIN_ROSTER_SERVICE_SERVER="http://127.0.0.1:5874"
export KURISU_SERVICE_SERVER="http://127.0.0.1:5874"
export LOGIN_SERVICE_SERVER="http://127.0.0.1:5874"

pnpm install

npx react-router dev --host
