#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

command="${1}"
focus="${2:-"main"}"

user_handle="$(git config user.email | cut -d'@' -f1)"
monorepo_home="$(dirname "$(git rev-parse --git-common-dir)")"

if [[ "${command}" == "push" ]]; then
  cat <<EOF | sudo tee /etc/cron.d/devcron-${command}-${focus}
10 * * * * ${USER} ${monorepo_home}/main/seed/devprod/devcron/push.sh "${monorepo_home}" "${user_handle}" "${focus}" >/tmp/devcron-${command}-${focus}.log 2>&1
EOF
elif [[ "${command}" == "commit" ]]; then
  cat <<EOF | sudo tee /etc/cron.d/devcron-${command}-${focus}
*/5 * * * * ${USER} ${monorepo_home}/main/seed/devprod/devcron/commit.sh "${monorepo_home}" "${user_handle}" "${focus}" >/tmp/devcron-${command}-${focus}.log 2>&1
EOF
else
  echo "Unknown command: ${command}"
  exit 1
fi
