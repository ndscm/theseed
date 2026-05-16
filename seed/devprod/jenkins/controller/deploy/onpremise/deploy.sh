#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../../.."

server=${1:-"jenkins-controller"}
service_user=${2:-"jenkins-controller"}
port=${3:-"8080"}

container_engine=${CONTAINER_ENGINE:-"podman"}
if [[ "${container_engine}" != "podman" ]]; then
  printf 'Only podman is supported for now\n'
  exit 1
fi

mount_jenkins_home=${MOUNT_JENKINS_HOME:-"/mnt/data/jenkins/controller/jenkins_home"}

./seed/devprod/jenkins/controller/container/build.sh

# Stage the container image on the remote.
podman save "ghcr.io/ndscm/seed-devprod-jenkins-controller-container:latest" | ssh "${server}" "cat > /tmp/seed-devprod-jenkins-controller-container.tar"

# Deploy quadlet unit, then start the service.
#
# Becareful to preserve the ~/.local/share/containers uid/gid ownerships, which is the subuid of the running container.
cat <<END | ssh "${server}" "cat > ~/install-jenkins-controller.sh"
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
state_dir=\${service_user_home}/.local/state/jenkins-controller
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p "\${quadlet_dir}"
sudo machinectl shell '${service_user}@' /usr/bin/mkdir -p "\${state_dir}"

cat <<EOF | sudo tee "\${quadlet_dir}/jenkins-controller.container"
[Unit]
Description=Jenkins Controller Container

[Container]
ContainerName=jenkins-controller
Image=ghcr.io/ndscm/seed-devprod-jenkins-controller-container:latest
PidsLimit=-1
RunInit=true
Network=host
EnvironmentFile=%S/jenkins-controller/env
Volume=${mount_jenkins_home}:/var/jenkins_home:U,Z

[Service]
Restart=always
RestartSec=2s
TasksMax=infinity

[Install]
WantedBy=default.target
EOF

cat <<EOF | sudo tee "\${state_dir}/env"
JENKINS_OPTS=--httpPort=${port}
EOF

echo 'Creating quadlet volumes...' >&2
sudo mkdir -p "${mount_jenkins_home}"
sudo chown '${service_user}:${service_user}' "${mount_jenkins_home}"

echo 'Loading container and restarting service...' >&2
sudo chmod 644 /tmp/seed-devprod-jenkins-controller-container.tar
sudo machinectl shell '${service_user}@' /bin/bash -c 'podman load < /tmp/seed-devprod-jenkins-controller-container.tar'
sudo machinectl shell '${service_user}@' /bin/bash -c 'systemctl --user daemon-reload && systemctl --user restart jenkins-controller'

echo 'Cleaning up...' >&2
rm -f /tmp/seed-devprod-jenkins-controller-container.tar
END

ssh -t "${server}" "
chmod +x ~/install-jenkins-controller.sh
~/install-jenkins-controller.sh
rm -f ~/install-jenkins-controller.sh
"

printf "Use this command to access the container:\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${service_user}@ /bin/bash\x1b[0m\n"
printf "    \x1b[1;33mssh -t ${server} sudo machinectl shell ${service_user}@ /usr/bin/podman exec --interactive --tty --user jenkins jenkins-controller /bin/bash\x1b[0m\n"
