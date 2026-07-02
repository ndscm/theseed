#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

server=${1:-"svn.ndscm.biz"}
service_user=${2:-"svn"}
port=${3:-"7863"} # SVND
https=${SVN_HTTPS:-"true"}

container_engine=${CONTAINER_ENGINE:-"podman"}
if [[ "${container_engine}" != "podman" ]]; then
  printf 'Only podman is supported for now\n'
  exit 1
fi

mount_var_svn=${MOUNT_VAR_SVN:-"/mnt/data/svn/var/svn"}

export CONTAINER_ENGINE="${container_engine}"
./seed/devprod/subversion/container/build.sh

# Stage the container image on the remote.
podman save "ghcr.io/ndscm/seed-devprod-subversion-container:latest" | ssh "${server}" "cat > /tmp/seed-devprod-subversion-container.tar"

# Deploy quadlet unit, then start the service.
#
# Becareful to preserve the ~/.local/share/containers uid/gid ownerships, which is the subuid of the running container.
cat <<END | ssh "${server}" "cat > ~/install-subversion.sh"
#!/usr/bin/env bash
set -eux
set -o pipefail

echo 'Ensuring service user exists...' >&2
if ! id '${service_user}' &>/dev/null; then
  sudo useradd --create-home --shell /usr/sbin/nologin '${service_user}'
  sudo loginctl enable-linger '${service_user}'
fi

echo 'Creating quadlets...' >&2
service_user_home=\$(eval echo ~'${service_user}')
quadlet_dir=\${service_user_home}/.config/containers/systemd
state_dir=\${service_user_home}/.local/state/subversion
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p "\${quadlet_dir}"
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p "\${state_dir}"

cat <<EOF | sudo tee "\${quadlet_dir}/subversion.container"
[Unit]
Description=Subversion (Apache mod_dav_svn) Container

[Container]
ContainerName=subversion
Image=ghcr.io/ndscm/seed-devprod-subversion-container:latest
PidsLimit=-1
RunInit=true
Network=host
EnvironmentFile=%S/subversion/env
Volume=${mount_var_svn}:/var/svn:Z

[Service]
Restart=always
RestartSec=2s
TasksMax=infinity

[Install]
WantedBy=default.target
EOF

cat <<EOF | sudo tee "\${state_dir}/env" >/dev/null
SVN_HTTPD_PORT=${port}
SVN_SERVER_NAME=${server}
SVN_HTTPS=${https}
EOF

echo 'Creating quadlet volumes...' >&2
sudo mkdir -p "${mount_var_svn}"
sudo chown '${service_user}:${service_user}' "${mount_var_svn}"

echo 'Loading container and restarting service...' >&2
sudo chmod 644 /tmp/seed-devprod-subversion-container.tar
sudo machinectl shell '${service_user}@' /bin/bash -c 'podman load < /tmp/seed-devprod-subversion-container.tar'
sudo machinectl shell '${service_user}@' /bin/bash -c 'systemctl --user daemon-reload && systemctl --user restart subversion'

echo 'Cleaning up...' >&2
rm -f /tmp/seed-devprod-subversion-container.tar
END

ssh -t "${server}" "
chmod +x ~/install-subversion.sh
~/install-subversion.sh
rm -f ~/install-subversion.sh
"

printf "Use these commands to access the container:\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${service_user}@ /bin/bash\x1b[0m\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${service_user}@ /usr/bin/podman exec --interactive --tty subversion /bin/bash\x1b[0m\n"
printf "Create a repository with:\n"
printf "    \x1b[1;33m... podman exec --user www-data subversion svnadmin create /var/svn/repos/<name>\x1b[0m\n"
printf "Add an HTTP user with:\n"
printf "    \x1b[1;33m... podman exec --user www-data subversion htpasswd -B /var/svn/conf/htpasswd <username>\x1b[0m\n"
