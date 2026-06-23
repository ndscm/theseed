#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",system,"* ]]; then
    team_handle="${1:-"ndscm"}"
    monodrive_server="${2:-"drive.ndscm.com"}"

    printf "\e[34m[system] Checking monodrive...\e[0m\n"

    if [[ "${oslike}" == "debian" ]]; then
        if [[ "${run_sudo}" == "true" ]]; then
            sudo apt install -y davfs2

            davfs2_uid=$(id -u davfs2)
            davfs2_gid=$(id -g davfs2)
            if [[ -z "${SEED_MONODRIVE_LOGIN+x}" ]]; then
                read -p "Enter mono drive user login: " SEED_MONODRIVE_LOGIN
            fi
            cat <<EOF >>"${HOME}/.managed_profile.tmp"

## Monodrive

export SEED_MONODRIVE_LOGIN="${SEED_MONODRIVE_LOGIN}"
EOF
            if ! sudo grep "^[/]mnt[/]${team_handle}[/]drive[ ]" /etc/davfs2/secrets; then
                read -p "Enter mono drive user password: " SEED_MONODRIVE_PASSWORD
                printf "\n/mnt/${team_handle}/drive ${SEED_MONODRIVE_LOGIN} ${SEED_MONODRIVE_PASSWORD}\n" | sudo tee -a /etc/davfs2/secrets
            fi
            sudo usermod -aG davfs2 "${USER}"
            sudo mkdir -p "/mnt/${team_handle}/drive"
            cat <<EOF | sudo tee "/usr/lib/systemd/system/mnt-${team_handle}-drive.mount"
[Unit]
Description=Mount Seed Monodrive
After=network-online.target remote-fs.target
Wants=network-online.target

[Mount]
What=https://${monodrive_server}/remote.php/dav/files/${SEED_MONODRIVE_LOGIN}/${team_handle}
Where=/mnt/${team_handle}/drive
Options=uid=${davfs2_uid},gid=${davfs2_gid},file_mode=0664,dir_mode=2775,grpid
Type=davfs
TimeoutSec=15

[Install]
WantedBy=multi-user.target
EOF
            sudo systemctl daemon-reload
            sudo systemctl enable "mnt-${team_handle}-drive.mount"
            sudo systemctl start "mnt-${team_handle}-drive.mount"
        fi
    fi

    printf "\e[32m[system] Check monodrive done.\e[0m\n"
fi
