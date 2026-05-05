#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

server=${1:-"steins.ndscm.com"}
mount_etc_steins=${2:-"/mnt/data/steins/etc/steins"}

container_cli=${CONTAINER_CLI:-"docker"}

./seed/newtype/hooin/container/build.sh
"${container_cli}" save ghcr.io/ndscm/seed-newtype-hooin-container:latest | ssh "${server}" "${container_cli} load"

printf "Use this command to access the container:\n"
printf "    \x1b[1;33m${container_cli} --host \"ssh://${server}\" exec --interactive --tty --user hooin hooin zsh\x1b[0m\n"

"${container_cli}" --host "ssh://${server}" rm -f hooin || true
"${container_cli}" --host "ssh://${server}" run --name hooin --interactive --tty \
  --network=host \
  --volume "${mount_etc_steins}:/etc/steins" \
  ghcr.io/ndscm/seed-newtype-hooin-container:latest
