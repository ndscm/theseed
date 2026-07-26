#!/usr/bin/env bash
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
  set +x
  age_key_file="${AGE_KEY_FILE:-""}"
  if [[ -z "${age_key_file}" ]]; then
    age_key_file="$(ndscm secret --user get-path key.age)"
  fi
  openid_client_secret="$(
    age -d \
      -i "${age_key_file}" \
      "$(ndscm secret --user get-path seed/newtype/kurisu/webapp/OPENID_CLIENT_SECRET.age)"
  )"
  exec {openid_client_secret_fd}< <(printf "%s" "${openid_client_secret}")
  unset openid_client_secret
  set -x
  export OPENID_CLIENT_SECRET_FILE="/proc/$$/fd/${openid_client_secret_fd}"
fi

npx react-router dev --host
