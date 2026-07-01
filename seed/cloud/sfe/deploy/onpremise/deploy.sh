#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

server=${1:-"sfe"}
service_user=${2:-"sfe"}
sfe_openid_client_id=${3:-""}
sfe_openid_client_secret=${4:-""}

container_engine=${CONTAINER_ENGINE:-"podman"}
if [[ "${container_engine}" != "podman" ]]; then
  printf 'Only podman is supported for now\n'
  exit 1
fi

export CONTAINER_ENGINE="${container_engine}"
./seed/cloud/sfe/container/build.sh

# Stage the container image on the remote.
podman save "ghcr.io/ndscm/seed-cloud-sfe-container:latest" | ssh "${server}" "cat > /tmp/seed-cloud-sfe-container.tar"

# Deploy quadlet unit, then start the service.
#
# Becareful to preserve the ~/.local/share/containers uid/gid ownerships, which is the subuid of the running container.
cat <<END | ssh "${server}" "cat > ~/install-sfe.sh"
#!/usr/bin/env bash
set -eux
set -o pipefail

echo 'Ensuring service user exists...' >&2
if ! id '${service_user}' &>/dev/null; then
  sudo useradd --create-home --shell /usr/sbin/nologin '${service_user}'
  sudo loginctl enable-linger '${service_user}'
fi

if [[ -n "${sfe_openid_client_secret}" ]]; then
  echo 'Creating openid client secret...' >&2
  set +x
  sudo machinectl shell '${service_user}@' /bin/bash -c 'printf %s "${sfe_openid_client_secret}" | podman secret create --replace SFE_OPENID_CLIENT_SECRET -'
  printf 'Created podman secret: %s\n' "SFE_OPENID_CLIENT_SECRET" >&2
  set -x
fi

echo 'Creating quadlets...' >&2
service_user_home=\$(eval echo ~'${service_user}')
quadlet_dir=\${service_user_home}/.config/containers/systemd
state_dir=\${service_user_home}/.local/state/sfe
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p "\${quadlet_dir}"
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p "\${state_dir}"

cat <<EOF | sudo tee "\${quadlet_dir}/sfe.container"
[Unit]
Description=SFE Server Container

[Container]
ContainerName=sfe
Image=ghcr.io/ndscm/seed-cloud-sfe-container:latest
PidsLimit=-1
RunInit=true
Network=host
Secret=SFE_OPENID_CLIENT_SECRET
EnvironmentFile=%S/sfe/env
Exec=--http=escalate --https route --verbose
${EXTRA_SFE_CONTAINER_CONFIG:-}

[Service]
Restart=always
RestartSec=2s
TasksMax=infinity

[Install]
WantedBy=default.target
EOF

cat <<EOF | sudo tee "\${state_dir}/env"
SFE_OPENID_DISCOVERY_URL=https://account.ndscm.com/realms/ndscm/.well-known/openid-configuration
SFE_OPENID_CLIENT_ID=${sfe_openid_client_id}
SFE_OPENID_CLIENT_SECRET_FILE=/run/secrets/SFE_OPENID_CLIENT_SECRET
SFE_SFE_CERTIFICATE_SERVICE_SERVER=https://certificate.sfe.ndscm.com
${EXTRA_SFE_ENV:-}
EOF

echo 'Loading container and restarting service...' >&2
sudo chmod 644 /tmp/seed-cloud-sfe-container.tar
sudo machinectl shell '${service_user}@' /bin/bash -c 'podman load < /tmp/seed-cloud-sfe-container.tar'
sudo machinectl shell '${service_user}@' /bin/bash -c 'systemctl --user daemon-reload && systemctl --user restart sfe'

echo 'Cleaning up...' >&2
rm -f /tmp/seed-cloud-sfe-container.tar
END

ssh -t "${server}" "
chmod +x ~/install-sfe.sh
~/install-sfe.sh
rm -f ~/install-sfe.sh
"

printf "Use this command to access the container:\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${service_user}@ /bin/bash\x1b[0m\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${service_user}@ /usr/bin/podman exec --interactive --tty --user sfe sfe /bin/bash\x1b[0m\n"

ssh -t "${server}" sudo machinectl shell "${service_user}@" /usr/bin/podman logs --follow sfe
