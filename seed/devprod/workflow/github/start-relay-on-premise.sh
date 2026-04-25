#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

server=${1:-"workflow.ndscm.com"}
container_name=${2:-"workflow-github-webhook-relay"}
event_source=${3:-"https://webhook.ndscm.com/github/subscribe"}
relay_to=${4:-"https://workflow.ndscm.com/generic-webhook-trigger/invoke"}

./seed/devprod/webhook/relay/container/build.sh
"${CONTAINER_CLI}" save ghcr.io/ndscm/seed-devprod-webhook-relay-container:latest | ssh "${server}" "${CONTAINER_CLI} load"

printf "Use this command to access the container:\n"
printf "    \x1b[1;33m${CONTAINER_CLI} --host \"ssh://${server}\" exec --interactive --tty --user relay ${container_name} bash\x1b[0m\n"

"${CONTAINER_CLI}" --host "ssh://${server}" rm -f ${container_name} || true
"${CONTAINER_CLI}" --host "ssh://${server}" run --name ${container_name} --interactive --tty \
  --network=host \
  --restart=always \
  ghcr.io/ndscm/seed-devprod-webhook-relay-container:latest \
  --event_source "${event_source}" \
  --relay_to "${relay_to}" \
  --verbose
