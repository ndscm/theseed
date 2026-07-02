#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  repo_identifier="${1:-"theseed"}"
  monodrive_server="${2:-"drive.ndscm.com"}"
  monodrive_directory="${3:-"theseed"}"
  mount_path="${ND_REPOS_HOME}/${repo_identifier}/drive/main"

  printf "\e[34m[user] Checking monodrive...\e[0m\n"

  if [[ "${oslike}" == "debian" ]]; then
    if ! rclone listremotes | grep -qx "${repo_identifier}:"; then
      read -p "Enter monodrive user login: " monodrive_login
      read -s -p "Enter monodrive user password: " monodrive_password
      printf "\n"
      set +x
      rclone config create "${repo_identifier}" webdav \
        url="https://${monodrive_server}/remote.php/dav/files/${monodrive_login}" \
        vendor=nextcloud \
        user="${monodrive_login}" \
        pass="${monodrive_password}" \
        --obscure
      set -x
    fi

    mkdir -p "${mount_path}"
    mkdir -p "${HOME}/.config/systemd/user"

    cat >"${HOME}/.config/systemd/user/ndscm-${repo_identifier}-drive-main.service" <<EOF
[Unit]
Description=Mount Seed Monodrive (rclone)
After=network-online.target
Wants=network-online.target

[Service]
Type=notify
ExecStart=/usr/bin/rclone mount "${repo_identifier}:${monodrive_directory}" "${mount_path}" --vfs-cache-mode writes
ExecStop=/usr/bin/fusermount3 -u "${mount_path}"
Restart=on-failure
RestartSec=10

[Install]
WantedBy=default.target
EOF

    systemctl --user daemon-reload
    systemctl --user enable --now "ndscm-${repo_identifier}-drive-main.service"
  fi

  printf "\e[32m[user] Check monodrive done.\e[0m\n"
fi
