#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

container_engine=${CONTAINER_ENGINE:-"podman"}
./seed/devprod/container/ubuntu/build.sh

bazel build //seed/devprod/golink/server:golink-server_tar_gz
cp -f ./bazel-bin/seed/devprod/golink/server/golink-server.tar.gz ./seed/devprod/golink/container/golink-server.tar.gz

cd ./seed/devprod/golink/container/

build_compat=()
if [[ "${container_engine}" == "docker" ]]; then
  build_compat+=("-f" "Containerfile")
elif [[ "${container_engine}" == "podman" ]]; then
  build_compat+=("--userns" "auto:size=65536")
fi

"${container_engine}" build "${build_compat[@]}" -t ghcr.io/ndscm/seed-devprod-golink-container:latest .
