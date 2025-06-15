#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

server=${1-"devbox.dev.${ND_USER_HANDLE}.pek.ndscm.com"}

scp ./maintain-managed-instance.sh ${server}:maintain-managed-instance.sh
ssh -t ${server} "
set -eux
sudo mv maintain-managed-instance.sh /usr/local/sbin/maintain-managed-instance.sh
maintain-managed-instance.sh
"
