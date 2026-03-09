#!/bin/bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../..

command=${1}
branch=${2}
nd_user_handle=${ND_USER_HANDLE}

if [[ "${command}" == "push" ]]; then
  cat <<EOF | sudo tee /etc/cron.d/devcron-${command}-${branch}
10 * * * * ${USER} ${HOME}/theseed/main/seed/devprod/devcron/push.sh ${branch} ${nd_user_handle} >/tmp/devcron-${command}-${branch}.log 2>&1
EOF
elif [[ "${command}" == "commit" ]]; then
  cat <<EOF | sudo tee /etc/cron.d/devcron-${command}-${branch}
*/5 * * * * ${USER} ${HOME}/theseed/main/seed/devprod/devcron/commit.sh ${branch} ${nd_user_handle} >/tmp/devcron-${command}-${branch}.log 2>&1
EOF
else
  echo "Unknown command: ${command}"
  exit 1
fi
