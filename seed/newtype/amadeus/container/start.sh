#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

user_handle="${1:-"christina"}"

container_engine="${CONTAINER_ENGINE:-"podman"}"
if [[ "${container_engine}" != "podman" ]]; then
  printf 'Only podman is supported\n'
  exit 1
fi

export CONTAINER_ENGINE="${container_engine}"
./build.sh

podman run \
  --name "${user_handle}" \
  --replace \
  --rm \
  --interactive \
  --tty \
  --pids-limit -1 \
  --network host \
  --device /dev/fuse \
  --privileged \
  --systemd always \
  --secret SILICON_REFRESH_TOKEN \
  --volume "/mnt/data/${user_handle}/playpen/home:/playpen/home:Z" \
  ghcr.io/ndscm/seed-newtype-amadeus-container:latest \
  /opt/amadeus/amadeus-server \
  --openid_discovery_url=https://account.ndscm.com/realms/ndscm/.well-known/openid-configuration \
  --playpen "${user_handle}" \
  --silicon_refresh_token_file=/run/secrets/SILICON_REFRESH_TOKEN \
  --verbose
