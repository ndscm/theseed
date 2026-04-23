#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

server=${1:-"cache.rbe.ndscm.com"}
mount_www=${2:-"/mnt/data/rbe/cache/var/www"}

./build.sh
"${CONTAINER_CLI}" save ghcr.io/ndscm/seed-devprod-rbe-cache:latest | ssh "${server}" "${CONTAINER_CLI} load"

"${CONTAINER_CLI}" --host "ssh://${server}" rm -f rbe-cache || true
"${CONTAINER_CLI}" --host "ssh://${server}" run --name rbe-cache --interactive --tty \
  --network=host \
  --volume "${mount_www}:/var/www" \
  --restart unless-stopped \
  ghcr.io/ndscm/seed-devprod-rbe-cache:latest
