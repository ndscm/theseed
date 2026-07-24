#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

stack=${1:-"local"}

pnpm install

export HOOIN_DICTATE_SERVICE_SERVER="http://127.0.0.1:5874"
export HOOIN_INVADE_SERVICE_SERVER="http://127.0.0.1:5874"
export HOOIN_RAID_SERVICE_SERVER="http://127.0.0.1:5874"
export HOOIN_ROSTER_SERVICE_SERVER="http://127.0.0.1:5874"
export KURISU_SERVICE_SERVER="http://127.0.0.1:5874"
export LOGIN_SERVICE_SERVER="http://127.0.0.1:5874"

export OPENID_CLIENT_ID="kurisu-webapp-dev"
if [[ -f "$(ndscm secret --user get-path seed/newtype/kurisu/webapp/OPENID_CLIENT_SECRET.age)" ]]; then
  export OPENID_CLIENT_ID="kurisu-webapp-dev-${ND_USER_HANDLE}"
  secret_file="$(mktemp)"
  exec {secret_fd}<>"${secret_file}"
  rm -f "${secret_file}"
  set +x
  age -d \
    -i "$(ndscm secret --user get-path key.age)" \
    "$(ndscm secret --user get-path seed/newtype/kurisu/webapp/OPENID_CLIENT_SECRET.age)" \
    >&"$secret_fd"
  set -x
  export OPENID_CLIENT_SECRET_FILE="/proc/$$/fd/$secret_fd"
fi

npx react-router dev --host
