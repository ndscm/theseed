#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

container_engine=${CONTAINER_ENGINE:-"podman"}

export CONTAINER_ENGINE="${container_engine}"
./seed/devprod/container/ubuntu/build.sh

mkdir -p ./seed/newtype/amadeus/container/bin/

bazel build //seed/devprod/ndscm/cli
cp -f ./bazel-bin/seed/devprod/ndscm/cli/ndscm_/ndscm ./seed/newtype/amadeus/container/bin/ndscm

cp -f ./seed/vendor/bazel/bin/bazel ./seed/newtype/amadeus/container/bin/bazel
cp -f ./seed/vendor/bazel/bin/bazelisk ./seed/newtype/amadeus/container/bin/bazelisk
cp -f ./seed/vendor/buck/bin/buck2 ./seed/newtype/amadeus/container/bin/buck2
cp -f ./seed/vendor/buck/bin/reindeer ./seed/newtype/amadeus/container/bin/reindeer
cp -f ./seed/vendor/claude/bin/claude ./seed/newtype/amadeus/container/bin/claude
cp -f ./seed/vendor/crane/bin/crane ./seed/newtype/amadeus/container/bin/crane
cp -f ./seed/vendor/lego/bin/lego ./seed/newtype/amadeus/container/bin/lego
cp -f ./seed/vendor/llvm/bin/clang ./seed/newtype/amadeus/container/bin/clang

bazel build //seed/newtype/amadeus/server
cp -f ./bazel-bin/seed/newtype/amadeus/server/amadeus-server_/amadeus-server ./seed/newtype/amadeus/container/amadeus-server

cd ./seed/newtype/amadeus/container/

build_compat=()
if [[ "${container_engine}" == "docker" ]]; then
  build_compat+=("-f" "Containerfile")
elif [[ "${container_engine}" == "podman" ]]; then
  build_compat+=("--userns" "auto:size=65536")
fi

"${container_engine}" build "${build_compat[@]}" -t ghcr.io/ndscm/seed-newtype-amadeus-container:latest .
