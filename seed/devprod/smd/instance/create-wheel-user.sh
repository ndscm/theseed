#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

server=${1-"devbox.dev.${ND_USER_HANDLE}.pek.ndscm.com"}
cloud=${2-"ndscm"}
user_nd_user_handle=${3-"${ND_USER_HANDLE}"}
user_authorized_key=${4-"$(cat ~/.ssh/id_ed25519.pub)"}

ssh -i ${ND_USER_SECRET_HOME}/${cloud}/${ND_USER_HANDLE}.pem -t root@${server} "
set -eux

if [[ -d /home/${user_nd_user_handle} ]]; then
  echo ${user_authorized_key} >>/home/${user_nd_user_handle}/.ssh/authorized_keys
  exit 0
fi

useradd --create-home ${user_nd_user_handle}
passwd ${user_nd_user_handle}
usermod -a -G wheel ${user_nd_user_handle}
mkdir /home/${user_nd_user_handle}/.ssh
chmod 0700 /home/${user_nd_user_handle}/.ssh
echo ${user_authorized_key} >/home/${user_nd_user_handle}/.ssh/authorized_keys
chmod 0600 /home/${user_nd_user_handle}/.ssh/authorized_keys
chown -R ${user_nd_user_handle}:${user_nd_user_handle} /home/${user_nd_user_handle}/.ssh
"
