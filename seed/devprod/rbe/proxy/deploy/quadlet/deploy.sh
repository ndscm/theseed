#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../../.."

server="${1:-"cache.rbe.ndscm.biz"}"
service_user="${2:-"rbe"}"
port="${3:-"22243"}"
remote_cache="${4:-"https://ndscm-prod-rbe-cache-prod.storage.googleapis.com/"}"
set +x
openid_client_secret="${5:-""}"
set -x

mount_cache=${MOUNT_CACHE:-"/mnt/data/rbe/cache"}

# Stage the container image on the remote.
bazel run --stamp \
  --@rules_img//img/settings:load_daemon=tar \
  //seed/devprod/rbe/proxy/deploy/quadlet:load |
  ssh "${server}" "cat > ~/seed-devprod-rbe-proxy-deploy-quadlet.tar"

# Deploy quadlet unit, then start the service.
#
# Be careful to preserve the ~/.local/share/containers uid/gid ownerships, which is the subuid of the running container.

printf 'Ensuring service user existence...\n' >&2
ssh -t "${server}" "#!/usr/bin/env bash
set -eux
set -o pipefail

if ! id '${service_user}' &>/dev/null; then
  sudo useradd --create-home --shell /usr/sbin/nologin '${service_user}'
  sudo loginctl enable-linger '${service_user}'
fi
"

set +x
if [[ -n "${openid_client_secret}" ]]; then
  printf 'Updating openid client secret...' >&2
  printf '%s' "${openid_client_secret}" | ssh "${server}" "cat >~/OPENID_CLIENT_SECRET"
  ssh -t "${server}" "#!/usr/bin/env bash
set -eux
set -o pipefail

sudo mv ~/OPENID_CLIENT_SECRET ~${service_user}/OPENID_CLIENT_SECRET
sudo chown '${service_user}:${service_user}' ~${service_user}/OPENID_CLIENT_SECRET
trap 'sudo rm -f ~${service_user}/OPENID_CLIENT_SECRET' EXIT
sudo machinectl shell '${service_user}@' /bin/bash -c 'podman secret create --replace OPENID_CLIENT_SECRET ~/OPENID_CLIENT_SECRET'
"
fi
set -x

cat <<END | ssh "${server}" "cat > ~/install-rbe-proxy.sh"
#!/usr/bin/env bash
set -eux
set -o pipefail

printf 'Loading container image...\n' >&2
sudo mv ~/seed-devprod-rbe-proxy-deploy-quadlet.tar ~${service_user}/seed-devprod-rbe-proxy-deploy-quadlet.tar
sudo chown '${service_user}:${service_user}' ~${service_user}/seed-devprod-rbe-proxy-deploy-quadlet.tar
trap 'sudo rm -f ~${service_user}/seed-devprod-rbe-proxy-deploy-quadlet.tar' EXIT
sudo machinectl shell '${service_user}@' /bin/bash -c 'podman load --input ~/seed-devprod-rbe-proxy-deploy-quadlet.tar'
sudo machinectl shell '${service_user}@' /bin/bash -c 'podman image exists ghcr.io/ndscm/seed-devprod-rbe-proxy-deploy-quadlet:latest'

printf 'Creating quadlets...\n' >&2
service_user_home=\$(eval printf ~'${service_user}')
quadlet_dir=\${service_user_home}/.config/containers/systemd
state_dir=\${service_user_home}/.local/state/rbe-proxy
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p "\${quadlet_dir}"
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p "\${state_dir}"

cat <<EOF | sudo tee "\${quadlet_dir}/rbe-proxy.container"
[Unit]
Description=RBE Proxy Container

[Container]
ContainerName=rbe-proxy
Image=ghcr.io/ndscm/seed-devprod-rbe-proxy-deploy-quadlet:latest
PidsLimit=-1
RunInit=true
Network=host
EnvironmentFile=%S/rbe-proxy/env
Secret=OPENID_CLIENT_SECRET
Volume=${mount_cache}:/var/cache/rbe/cache:U,Z

[Service]
Restart=always
RestartSec=2s
TasksMax=infinity

[Install]
WantedBy=default.target
EOF

cat <<EOF | sudo tee "\${state_dir}/env"
RBE_PROXY_REMOTE_CACHE=${remote_cache}
RBE_PROXY_LOCAL_CACHE=/var/cache/rbe/cache
RBE_PROXY_LOCAL_PUT=true
RBE_PROXY_LISTEN=:${port}
RBE_PROXY_CREDENTIAL_HELPER=/opt/rbe/rbe-credential-helper.sh
RBE_PROXY_CREDENTIAL_HELPER_FORMAT=oauth2
EOF

printf 'Creating quadlet volumes...\n' >&2
sudo mkdir -p "${mount_cache}"
sudo chown '${service_user}:${service_user}' "${mount_cache}"

printf 'Restarting service...\n' >&2
sudo machinectl shell '${service_user}@' /bin/bash -c 'systemctl --user daemon-reload'
sudo machinectl shell '${service_user}@' /bin/bash -c 'systemctl --user restart rbe-proxy'
END

ssh -t "${server}" "
trap 'rm -f ~/install-rbe-proxy.sh' EXIT
chmod +x ~/install-rbe-proxy.sh
~/install-rbe-proxy.sh
"

printf "Use this command to login the quadlet user:\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${service_user}@ /bin/bash\x1b[0m\n"
printf "Use this command to access the container:\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${service_user}@ /usr/bin/podman exec -i -t rbe-proxy /bin/bash\x1b[0m\n"

ssh -t ${server} sudo machinectl shell ${service_user}@ /usr/bin/podman logs --follow rbe-proxy
