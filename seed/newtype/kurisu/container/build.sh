#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

container_engine=${CONTAINER_ENGINE:-"podman"}

export CONTAINER_ENGINE="${container_engine}"
./seed/devprod/container/ubuntu/build.sh

bazel build --stamp //seed/newtype/kurisu/server:kurisu-server_tar_gz
cp -f ./bazel-bin/seed/newtype/kurisu/server/kurisu-server.tar.gz ./seed/newtype/kurisu/container/kurisu-server.tar.gz

cd ./seed/newtype/kurisu/container/

build_compat=()
if [[ "${container_engine}" == "docker" ]]; then
  build_compat+=("-f" "Containerfile")
elif [[ "${container_engine}" == "podman" ]]; then
  build_compat+=("--userns" "auto:size=65536")
fi

"${container_engine}" build "${build_compat[@]}" -t ghcr.io/ndscm/seed-newtype-kurisu-container:latest .
