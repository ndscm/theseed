#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

container_engine=${CONTAINER_ENGINE:-"podman"}

export CONTAINER_ENGINE="${container_engine}"
./seed/devprod/container/debian/build.sh

bazel build //seed/cloud/sfe/certificate/server
cp -f ./bazel-bin/seed/cloud/sfe/certificate/server/sfe-certificate-server_/sfe-certificate-server \
  ./seed/cloud/sfe/certificate/container/sfe-certificate-server

cd ./seed/cloud/sfe/certificate/container/

build_compat=()
if [[ "${container_engine}" == "docker" ]]; then
  build_compat+=("-f" "Containerfile")
elif [[ "${container_engine}" == "podman" ]]; then
  build_compat+=("--userns" "auto:size=65536")
fi

"${container_engine}" build "${build_compat[@]}" -t ghcr.io/ndscm/seed-cloud-sfe-certificate-container:latest .
