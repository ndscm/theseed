#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

server=${1:-"newtype.ndscm.com"}
mount_etc_steins=${2:-"/mnt/data/steins/etc/steins"}
mount_home=${3:-"/mnt/data/christina/home"}

./build.sh
"${CONTAINER_CLI}" save ghcr.io/ndscm/seed-newtype-amadeus-container:latest | ssh "${server}" "${CONTAINER_CLI} load"

printf "Use this command to access the container:\n"
printf "    \x1b[1;33m${CONTAINER_CLI} --host \"ssh://${server}\" exec --interactive --tty --user amadeus christina zsh\x1b[0m\n"

"${CONTAINER_CLI}" --host "ssh://${server}" rm -f christina || true
"${CONTAINER_CLI}" --host "ssh://${server}" run --name christina --interactive --tty \
  --network=host \
  --volume "${mount_etc_steins}:/etc/steins" \
  --volume "${mount_home}:/home" \
  ghcr.io/ndscm/seed-newtype-amadeus-container:latest
