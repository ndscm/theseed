#!/bin/bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../..

branch=${1}
nd_user_handle=${ND_USER_HANDLE}

cat <<EOF | sudo tee /etc/cron.d/devcron-${branch}
10 * * * * ${USER} ${HOME}/theseed/main/seed/devprod/devcron/push.sh ${branch} ${nd_user_handle} >/tmp/devcron-${branch}.log 2>&1
EOF
