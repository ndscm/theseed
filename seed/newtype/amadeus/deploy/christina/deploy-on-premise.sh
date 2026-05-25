#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

server=${1:-"steins.ndscm.com"}
mount_etc_steins=${2:-"/mnt/data/steins/etc/steins"}
mount_home=${3:-"/mnt/data/christina/home"}

export CONTAINER_ENGINE="docker"
./seed/newtype/amadeus/container/build.sh

docker save ghcr.io/ndscm/seed-newtype-amadeus-container:latest | ssh "${server}" "docker load"

printf "Use this command to access the container:\n"
printf "    \x1b[1;33mdocker --host \"ssh://${server}\" exec --interactive --tty --user amadeus christina zsh\x1b[0m\n"

docker --host "ssh://${server}" rm -f christina || true
docker --host "ssh://${server}" run --name christina --interactive --tty \
  --network=host \
  --volume "${mount_etc_steins}:/etc/steins" \
  --volume "${mount_home}:/home" \
  ghcr.io/ndscm/seed-newtype-amadeus-container:latest \
  "/opt/amadeus/amadeus-server" \
  --port 2447 \
  --trust_openid_issuers_file /etc/steins/openid_issuers.json \
  --verbose
