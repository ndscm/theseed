#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

server=${1-"devbox.dev.${ND_USER_HANDLE}.pek.ndscm.com"}
cloud=${2-"ndscm"}

scp -i ${ND_USER_SECRET_HOME}/${cloud}/${ND_USER_HANDLE}.pem ./setup-managed-instance.sh root@${server}:/usr/local/sbin/setup-managed-instance.sh
ssh -i ${ND_USER_SECRET_HOME}/${cloud}/${ND_USER_HANDLE}.pem -t root@${server} "
set -eux
setup-managed-instance.sh
"

./create-remote-wheel-user.sh "${server}" "${cloud}" "${ND_USER_HANDLE}" "$(cat ~/.ssh/id_ed25519.pub)"

./maintain-remote-instance.sh "${server}"
