#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

server="${1:-"steins.ndscm.biz"}"
user_handle="${2:-"christina"}"
port="${3:-"2447"}"
hooin_url="${4:-"https://hooin.ndscm.biz/"}"
set +x
user_refresh_token="${5:-""}"
set -x

container_engine="${CONTAINER_ENGINE:-"podman"}"
if [[ "${container_engine}" != "podman" ]]; then
  printf 'Only podman is supported\n'
  exit 1
fi

mount_home=${MOUNT_HOME:-"/mnt/data/${user_handle}/home"}

export CONTAINER_ENGINE="${container_engine}"
./seed/newtype/amadeus/container/build.sh

# Stage the container image on the remote.
podman save "ghcr.io/ndscm/seed-newtype-amadeus-container:latest" | ssh "${server}" "cat > /tmp/seed-newtype-amadeus-container.tar"

# Deploy quadlet unit and the silicon refresh token secret, then start the service.
#
# Each silicon user runs as its own rootless service user, so the quadlet and the
# install script are namespaced by user_handle to coexist on a shared host.
#
# Becareful to preserve the ~/.local/share/containers uid/gid ownerships, which is the subuid of the running container.

printf 'Ensuring service user existence...\n' >&2
ssh -t "${server}" "#!/usr/bin/env bash
set -eux
set -o pipefail

if ! id '${user_handle}' &>/dev/null; then
  sudo useradd --create-home --shell /usr/sbin/nologin '${user_handle}'
  sudo loginctl enable-linger '${user_handle}'
fi
"

set +x
if [[ -n "${user_refresh_token}" ]]; then
  printf 'Updating silicon refresh token secret...' >&2
  printf '%s' "${user_refresh_token}" | ssh "${server}" "cat >~/SILICON_REFRESH_TOKEN"
  ssh -t "${server}" "#!/usr/bin/env bash
set -eux
set -o pipefail

sudo mv ~/SILICON_REFRESH_TOKEN ~${user_handle}/SILICON_REFRESH_TOKEN
sudo chown '${user_handle}:${user_handle}' ~${user_handle}/SILICON_REFRESH_TOKEN
trap 'sudo rm -f ~${user_handle}/SILICON_REFRESH_TOKEN' EXIT
sudo machinectl shell '${user_handle}@' /bin/bash -c 'podman secret create --replace SILICON_REFRESH_TOKEN ~/SILICON_REFRESH_TOKEN'
"
fi
set -x

cat <<END | ssh "${server}" "cat > ~/install-amadeus-${user_handle}.sh"
#!/usr/bin/env bash
set -eux
set -o pipefail

printf 'Creating quadlets...\n' >&2
service_user_home=\$(eval printf ~'${user_handle}')
quadlet_dir=\${service_user_home}/.config/containers/systemd
state_dir=\${service_user_home}/.local/state/amadeus
sudo machinectl shell '${user_handle}@' /usr/bin/mkdir -p "\${quadlet_dir}"
sudo machinectl shell '${user_handle}@' /usr/bin/mkdir -p "\${state_dir}"

cat <<EOF | sudo tee "\${quadlet_dir}/amadeus.container"
[Unit]
Description=Amadeus Server Container

[Container]
ContainerName=amadeus
Image=ghcr.io/ndscm/seed-newtype-amadeus-container:latest
Pull=never
PidsLimit=-1
RunInit=true
Network=host
Secret=SILICON_REFRESH_TOKEN
Volume=${mount_home}:/home:Z
EnvironmentFile=%S/amadeus/env
Exec=/opt/amadeus/amadeus-server --verbose

[Service]
Restart=always
RestartSec=2s
TasksMax=infinity

[Install]
WantedBy=default.target
EOF

cat <<EOF | sudo tee "\${state_dir}/env"
AMADEUS_PORT=${port}
AMADEUS_OPENID_DISCOVERY_URL=https://account.ndscm.com/realms/ndscm/.well-known/openid-configuration
AMADEUS_SILICON_REFRESH_TOKEN_FILE=/run/secrets/SILICON_REFRESH_TOKEN
AMADEUS_HOOIN_COMMUTE_SERVICE_SERVER=${hooin_url}
EOF

printf 'Creating quadlet volumes...\n' >&2
sudo mkdir -p "${mount_home}"
sudo chown '${user_handle}:${user_handle}' "${mount_home}"

printf 'Loading container and restarting service...\n' >&2
sudo chmod 644 /tmp/seed-newtype-amadeus-container.tar
sudo machinectl shell '${user_handle}@' /bin/bash -c 'podman load --input /tmp/seed-newtype-amadeus-container.tar'
sudo machinectl shell '${user_handle}@' /bin/bash -c 'podman image exists ghcr.io/ndscm/seed-newtype-amadeus-container:latest'
sudo machinectl shell '${user_handle}@' /bin/bash -c 'systemctl --user daemon-reload && systemctl --user restart amadeus'
END

ssh -t "${server}" "
trap 'rm -f ~/install-amadeus-${user_handle}.sh' EXIT
chmod +x ~/install-amadeus-${user_handle}.sh
~/install-amadeus-${user_handle}.sh
"

printf "Use this command to access the container:\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${user_handle}@ /bin/bash\x1b[0m\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${user_handle}@ /usr/bin/podman exec --interactive --tty --user amadeus amadeus /usr/bin/zsh\x1b[0m\n"

ssh -t ${server} sudo machinectl shell ${user_handle}@ /usr/bin/podman logs --follow amadeus
