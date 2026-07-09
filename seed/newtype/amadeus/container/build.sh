#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

container_engine=${CONTAINER_ENGINE:-"podman"}

export CONTAINER_ENGINE="${container_engine}"
./seed/devprod/container/ubuntu/build.sh

mkdir -p ./seed/newtype/amadeus/container/amadeus/
./seed/newtype/amadeus/playpen/container/build.sh
"${container_engine}" save "ghcr.io/ndscm/seed-newtype-amadeus-playpen-container:latest" \
  >"./seed/newtype/amadeus/container/amadeus/playpen.tar"
bazel build --stamp //seed/newtype/amadeus/server
cp -f ./bazel-bin/seed/newtype/amadeus/server/amadeus-server_/amadeus-server ./seed/newtype/amadeus/container/amadeus/amadeus-server

cd ./seed/newtype/amadeus/container/

build_compat=()
if [[ "${container_engine}" == "docker" ]]; then
  build_compat+=("-f" "Containerfile")
elif [[ "${container_engine}" == "podman" ]]; then
  build_compat+=("--userns" "auto:size=65536")
fi

"${container_engine}" build "${build_compat[@]}" -t ghcr.io/ndscm/seed-newtype-amadeus-container:latest .
