#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

server=${1:-"steins.ndscm.com"}
mount_etc_steins=${2:-"/mnt/data/steins/etc/steins"}

container_engine=${CONTAINER_ENGINE:-"docker"}
if [[ "${container_engine}" != "docker" ]]; then
  printf 'Only docker is supported for now\n'
  exit 1
fi

export CONTAINER_ENGINE="${container_engine}"
./seed/newtype/hooin/container/build.sh

docker save ghcr.io/ndscm/seed-newtype-hooin-container:latest | ssh "${server}" "docker load"

printf "Use this command to access the container:\n"
printf "    \x1b[1;33mdocker --host \"ssh://${server}\" exec --interactive --tty --user hooin hooin zsh\x1b[0m\n"

docker --host "ssh://${server}" rm -f hooin || true
docker --host "ssh://${server}" run --name hooin --interactive --tty \
  --network=host \
  --volume "${mount_etc_steins}:/etc/steins" \
  ghcr.io/ndscm/seed-newtype-hooin-container:latest
