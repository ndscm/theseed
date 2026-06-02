#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

server=${1:-"steins.ndscm.biz"}
service_user=${2:-"kurisu"}

container_engine=${CONTAINER_ENGINE:-"podman"}
if [[ "${container_engine}" != "podman" ]]; then
  printf 'Only podman is supported for now\n'
  exit 1
fi

export CONTAINER_ENGINE="${container_engine}"
./seed/newtype/kurisu/container/build.sh

# Stage the container image on the remote.
podman save "ghcr.io/ndscm/seed-newtype-kurisu-container:latest" | ssh "${server}" "cat > /tmp/seed-newtype-kurisu-container.tar"

# Deploy quadlet unit, then start the service.
#
# Becareful to preserve the ~/.local/share/containers uid/gid ownerships, which is the subuid of the running container.
cat <<END | ssh "${server}" "cat > ~/install-kurisu.sh"
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
state_dir=\${service_user_home}/.local/state/kurisu
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p "\${quadlet_dir}"
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p "\${state_dir}"

cat <<EOF | sudo tee "\${quadlet_dir}/kurisu.container"
[Unit]
Description=Kurisu Server Container

[Container]
ContainerName=kurisu
Image=ghcr.io/ndscm/seed-newtype-kurisu-container:latest
PidsLimit=-1
RunInit=true
Network=host
EnvironmentFile=%S/kurisu/env
Exec=--verbose

[Service]
Restart=always
RestartSec=2s
TasksMax=infinity

[Install]
WantedBy=default.target
EOF

cat <<EOF | sudo tee "\${state_dir}/env"
SEED_OPENID_DISCOVERY_URL=${SEED_OPENID_DISCOVERY_URL}
EOF

echo 'Loading container and restarting service...' >&2
sudo chmod 644 /tmp/seed-newtype-kurisu-container.tar
sudo machinectl shell '${service_user}@' /bin/bash -c 'podman load < /tmp/seed-newtype-kurisu-container.tar'
sudo machinectl shell '${service_user}@' /bin/bash -c 'systemctl --user daemon-reload && systemctl --user restart kurisu'

echo 'Cleaning up...' >&2
rm -f /tmp/seed-newtype-kurisu-container.tar
END

ssh -t "${server}" "
chmod +x ~/install-kurisu.sh
~/install-kurisu.sh
rm -f ~/install-kurisu.sh
"

printf "Use this command to access the container:\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${service_user}@ /bin/bash\x1b[0m\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${service_user}@ /usr/bin/podman exec --interactive --tty kurisu /bin/bash\x1b[0m\n"
