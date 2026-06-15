#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

server=${1:-"steins.ndscm.biz"}
user_handle=${2:-"christina"}
port=${3:-"2447"}
user_password=${4:-""}

container_engine=${CONTAINER_ENGINE:-"docker"}
if [[ "${container_engine}" != "docker" ]]; then
  printf 'Only docker is supported for now\n'
  exit 1
fi

mount_home=${MOUNT_HOME:-"/mnt/data/${user_handle}/home"}
mount_run_secrets=${MOUNT_RUN_SECRETS:-"/mnt/data/${user_handle}/run/secrets"}

export CONTAINER_ENGINE="${container_engine}"
./seed/newtype/amadeus/container/build.sh

if [[ -n "${user_password}" ]]; then
  ssh -t "${server}" "
if [[ ! -d \"${mount_run_secrets}\" ]]; then
  sudo mkdir -p \"${mount_run_secrets}\"
fi
"
  refresh_token=$(
    curl \
      -X POST \
      -L "https://account.ndscm.com/realms/ndscm/protocol/openid-connect/token" \
      -H "Content-Type: application/x-www-form-urlencoded" \
      -d "grant_type=password" \
      -d "client_id=silicon-prod" \
      -d "username=${user_handle}" \
      -d "password=${user_password}" \
      -d "scope=openid basic profile email offline_access" |
      jq -r ".refresh_token"
  )
  ssh -t "${server}" "
printf '${refresh_token}' | sudo tee \"${mount_run_secrets}/SILICON_REFRESH_TOKEN\" >/dev/null
sudo chmod 600 \"${mount_run_secrets}/SILICON_REFRESH_TOKEN\"
sudo chown 1001:1001 \"${mount_run_secrets}/SILICON_REFRESH_TOKEN\"
"
fi

docker save ghcr.io/ndscm/seed-newtype-amadeus-container:latest | ssh "${server}" "docker load"

printf "Use this command to access the container:\n"
printf "    \x1b[1;33mdocker --host \"ssh://${server}\" exec --interactive --tty --user amadeus ${user_handle} zsh\x1b[0m\n"

docker --host "ssh://${server}" rm -f "${user_handle}" || true
docker --host "ssh://${server}" run --name "${user_handle}" --interactive --tty \
  --network=host \
  --volume "${mount_home}:/home" \
  --volume "${mount_run_secrets}:/run/secrets" \
  ghcr.io/ndscm/seed-newtype-amadeus-container:latest \
  "/opt/amadeus/amadeus-server" \
  --port "${port}" \
  --openid_discovery_url "https://account.ndscm.com/realms/ndscm/.well-known/openid-configuration" \
  --silicon_refresh_token_file "/run/secrets/SILICON_REFRESH_TOKEN" \
  --hooin_commute_service_server "https://hooin.ndscm.biz" \
  --verbose
