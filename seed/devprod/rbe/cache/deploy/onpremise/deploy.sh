#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../../.."

server=${1:-"cache.rbe.ndscm.com"}
service_user=${2:-"rbe-cache"}

container_engine=${CONTAINER_ENGINE:-"podman"}
if [[ "${container_engine}" != "podman" ]]; then
  printf 'Only podman is supported for now\n'
  exit 1
fi

mount_www=${MOUNT_WWW:-"/mnt/data/rbe/cache/var/www"}

export CONTAINER_ENGINE="${container_engine}"
./seed/devprod/rbe/cache/container/build.sh

# Stage the container image on the remote.
podman save "ghcr.io/ndscm/seed-devprod-rbe-cache-container:latest" | ssh "${server}" "cat > /tmp/seed-devprod-rbe-cache-container.tar"

# Deploy quadlet unit, then start the service.
#
# Becareful to preserve the ~/.local/share/containers uid/gid ownerships, which is the subuid of the running container.
cat <<END | ssh "${server}" "cat > ~/install-rbe-cache.sh"
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
state_dir=\${service_user_home}/.local/state/rbe-cache
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p "\${quadlet_dir}"
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p "\${state_dir}"

cat <<EOF | sudo tee "\${quadlet_dir}/rbe-cache.container"
[Unit]
Description=RBE Cache Container

[Container]
ContainerName=rbe-cache
Image=ghcr.io/ndscm/seed-devprod-rbe-cache-container:latest
PidsLimit=-1
RunInit=true
Network=host
EnvironmentFile=%S/rbe-cache/env
Volume=${mount_www}:/var/www:U,Z

[Service]
Restart=always
RestartSec=2s
TasksMax=infinity

[Install]
WantedBy=default.target
EOF

cat <<EOF | sudo tee "\${state_dir}/env"
EOF

echo 'Creating quadlet volumes...' >&2
sudo mkdir -p "${mount_www}"
sudo chown '${service_user}:${service_user}' "${mount_www}"

echo 'Loading container and restarting service...' >&2
sudo chmod 644 /tmp/seed-devprod-rbe-cache-container.tar
sudo machinectl shell '${service_user}@' /bin/bash -c 'podman load < /tmp/seed-devprod-rbe-cache-container.tar'
sudo machinectl shell '${service_user}@' /bin/bash -c 'systemctl --user daemon-reload && systemctl --user restart rbe-cache'

echo 'Cleaning up...' >&2
rm -f /tmp/seed-devprod-rbe-cache-container.tar
END

ssh -t "${server}" "
chmod +x ~/install-rbe-cache.sh
~/install-rbe-cache.sh
rm -f ~/install-rbe-cache.sh
"

printf "Use this command to access the container:\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${service_user}@ /bin/bash\x1b[0m\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${service_user}@ /usr/bin/podman exec --interactive --tty rbe-cache /bin/bash\x1b[0m\n"
