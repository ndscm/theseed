#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

printf "\e[34mChecking theseed monodrive...\e[0m\n"

if [[ "${oslike}" == "debian" ]]; then
  davfs2_uid=$(id -u davfs2)
  davfs2_gid=$(id -g davfs2)
  if [[ -z "${SEED_MONODRIVE_LOGIN+x}" ]]; then
    read -p "Enter mono drive user login: " SEED_MONODRIVE_LOGIN
    cat <<EOF >>${HOME}/.seed_managed_profile.tmp

## Monodrive

export SEED_MONODRIVE_LOGIN="${SEED_MONODRIVE_LOGIN}"
EOF
  fi
  if ! sudo grep '^[/]mnt[/]theseed[/]drive[ ]' /etc/davfs2/secrets; then
    read -p "Enter mono drive user password: " SEED_MONODRIVE_PASSWORD
    printf "\n/mnt/theseed/drive ${SEED_MONODRIVE_LOGIN} ${SEED_MONODRIVE_PASSWORD}\n" | sudo tee -a /etc/davfs2/secrets
  fi
  sudo usermod -aG davfs2 ${USER}
  sudo mkdir -p /mnt/theseed/drive
  cat <<EOF | sudo tee /usr/lib/systemd/system/mnt-theseed-drive.mount
[Unit]
Description=Mount Theseed Monodrive
After=network-online.target remote-fs.target
Wants=network-online.target

[Mount]
What=https://drive.ndscm.com/remote.php/dav/files/${SEED_MONODRIVE_LOGIN}/theseed
Where=/mnt/theseed/drive
Options=uid=${davfs2_uid},gid=${davfs2_gid},file_mode=0664,dir_mode=2775,grpid
Type=davfs
TimeoutSec=15

[Install]
WantedBy=multi-user.target
EOF
  sudo systemctl daemon-reload
  sudo systemctl enable mnt-theseed-drive.mount
  sudo systemctl start mnt-theseed-drive.mount

  ln -s -f -n ${ND_MONOREPO_SECRET_HOME}/${ND_USER_HANDLE} ${ND_USER_SECRET_HOME}
fi

printf "\e[34mCheck theseed monodrive done.\e[0m\n"
