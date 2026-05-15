#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../../.."

server=${1:-"jenkins-node-1"}
service_user=${2:-"jenkins-node"}
jenkins_url=${3:-"https://workflow.ndscm.com/"}
jenkins_agent_name=${4:-"jenkins-node-1"}
set +x
jenkins_agent_secret=${5:-""}
printf 'jenkins_agent_secret=%s\n' "${jenkins_agent_secret:+REDACTED}" >&2
set -x

container_engine=${CONTAINER_ENGINE:-"podman"}
if [[ "${container_engine}" != "podman" ]]; then
  printf 'Only podman is supported for now\n'
  exit 1
fi

./seed/devprod/jenkins/node/container/build.sh

# Stage the container image on the remote.
podman save "ghcr.io/ndscm/seed-devprod-jenkins-node-container:latest" | ssh "${server}" "cat > /tmp/seed-devprod-jenkins-node-container.tar"

# Deploy quadlet unit and environment file, then start the service.
#
# Becareful to preserve the ~/.local/share/containers uid/gid ownerships, which is the subuid of the running container.
cat <<END | ssh "${server}" "cat > ~/install-jenkins-node.sh"
#!/usr/bin/env bash
set -eux
set -o pipefail

echo 'Ensuring service user exists...' >&2
if ! id '${service_user}' &>/dev/null; then
  sudo useradd --create-home --shell /usr/sbin/nologin '${service_user}'
  sudo loginctl enable-linger '${service_user}'
fi

if [[ -n "${jenkins_agent_secret}" ]]; then
  echo 'Creating Jenkins agent secret...' >&2
  set +x
  sudo machinectl shell '${service_user}@' /bin/bash -c 'printf "${jenkins_agent_secret}" | podman secret create --replace jenkins_agent_secret -'
  printf 'Created podman secret: %s\n' "jenkins_agent_secret" >&2
  set -x
fi

echo 'Creating quadlets...' >&2
service_user_home=\$(eval echo ~'${service_user}')
quadlet_dir=\${service_user_home}/.config/containers/systemd
state_dir=\${service_user_home}/.local/state/jenkins-node
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p "\${quadlet_dir}"
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p "\${state_dir}"

cat <<EOF | sudo tee "\${quadlet_dir}/jenkins-node.container"
[Unit]
Description=Jenkins Node Container

[Container]
ContainerName=jenkins-node
Image=ghcr.io/ndscm/seed-devprod-jenkins-node-container:latest
PidsLimit=-1
RunInit=true
Network=host
EnvironmentFile=%S/jenkins-node/env
Secret=jenkins_agent_secret
Volume=/mnt/data/jenkins/node/config/git:/home/jenkins/.config/git:ro,Z
Volume=/mnt/volatile/jenkins/node/agent:/home/jenkins/agent:U,Z
Volume=/mnt/volatile/jenkins/node/cache:/home/jenkins/.cache:U,Z

[Service]
Restart=on-failure
RestartSec=2s
SuccessExitStatus=143
TasksMax=infinity

[Install]
WantedBy=default.target
EOF

cat <<EOF | sudo tee "\${state_dir}/env"
JENKINS_URL=${jenkins_url}
JENKINS_WEB_SOCKET=true
JENKINS_AGENT_NAME=${jenkins_agent_name}
JENKINS_SECRET=@/run/secrets/jenkins_agent_secret
EOF

echo 'Creating quadlet volumes...' >&2
sudo mkdir -p /mnt/data/jenkins/node
sudo mkdir -p /mnt/volatile/jenkins/node
sudo chown '${service_user}:${service_user}' /mnt/data/jenkins/node
sudo chown '${service_user}:${service_user}' /mnt/volatile/jenkins/node
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p /mnt/data/jenkins/node/config/git
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p /mnt/volatile/jenkins/node/agent
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p /mnt/volatile/jenkins/node/cache

echo 'Loading container and restarting service...' >&2
sudo chmod 644 /tmp/seed-devprod-jenkins-node-container.tar
sudo machinectl shell '${service_user}@' /bin/bash -c 'podman load < /tmp/seed-devprod-jenkins-node-container.tar'
sudo machinectl shell '${service_user}@' /bin/bash -c 'systemctl --user daemon-reload && systemctl --user restart jenkins-node'

echo 'Cleaning up...' >&2
rm -f /tmp/seed-devprod-jenkins-node-container.tar
END

ssh -t "${server}" "
chmod +x ~/install-jenkins-node.sh
~/install-jenkins-node.sh
rm -f ~/install-jenkins-node.sh
"

printf "Use this command to access the container:\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${service_user}@ /bin/bash\x1b[0m\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${service_user}@ /usr/bin/podman exec --interactive --tty --user jenkins jenkins-node /bin/bash\x1b[0m\n"
