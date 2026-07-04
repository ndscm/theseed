#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

container_engine=${CONTAINER_ENGINE:-"podman"}

export CONTAINER_ENGINE="${container_engine}"
./seed/devprod/container/ubuntu/build.sh

mkdir -p ./seed/newtype/amadeus/container/bin/
cp -f ./seed/vendor/bazel/bin/bazel ./seed/newtype/amadeus/container/bin/bazel
cp -f ./seed/vendor/bazel/bin/bazelisk ./seed/newtype/amadeus/container/bin/bazelisk
cp -f ./seed/vendor/buck/bin/buck2 ./seed/newtype/amadeus/container/bin/buck2
cp -f ./seed/vendor/buck/bin/reindeer ./seed/newtype/amadeus/container/bin/reindeer
cp -f ./seed/vendor/claude/bin/claude ./seed/newtype/amadeus/container/bin/claude
cp -f ./seed/vendor/crane/bin/crane ./seed/newtype/amadeus/container/bin/crane
cp -f ./seed/vendor/docker/bin/docker ./seed/newtype/amadeus/container/bin/docker
cp -f ./seed/vendor/github/bin/gh ./seed/newtype/amadeus/container/bin/gh
cp -f ./seed/vendor/jq/bin/jq ./seed/newtype/amadeus/container/bin/jq
cp -f ./seed/vendor/lego/bin/lego ./seed/newtype/amadeus/container/bin/lego
cp -f ./seed/vendor/llvm/bin/clang ./seed/newtype/amadeus/container/bin/clang
cp -f ./seed/vendor/llvm/bin/clang-format ./seed/newtype/amadeus/container/bin/clang-format
cp -f ./seed/vendor/llvm/bin/clangd ./seed/newtype/amadeus/container/bin/clangd
cp -f ./seed/vendor/ndscm/bin/ndscm ./seed/newtype/amadeus/container/bin/ndscm
cp -f ./seed/vendor/node/bin/corepack ./seed/newtype/amadeus/container/bin/corepack
cp -f ./seed/vendor/node/bin/node ./seed/newtype/amadeus/container/bin/node
cp -f ./seed/vendor/node/bin/npm ./seed/newtype/amadeus/container/bin/npm
cp -f ./seed/vendor/node/bin/npx ./seed/newtype/amadeus/container/bin/npx
cp -f ./seed/vendor/node/bin/pnpm ./seed/newtype/amadeus/container/bin/pnpm
cp -f ./seed/vendor/node/bin/pnpx ./seed/newtype/amadeus/container/bin/pnpx
cp -f ./seed/vendor/node/bin/yarn ./seed/newtype/amadeus/container/bin/yarn
cp -f ./seed/vendor/node/bin/yarnpkg ./seed/newtype/amadeus/container/bin/yarnpkg
cp -f ./seed/vendor/podman/bin/podman-remote ./seed/newtype/amadeus/container/bin/podman-remote
cp -f ./seed/vendor/uv/bin/uv ./seed/newtype/amadeus/container/bin/uv
cp -f ./seed/vendor/uv/bin/uvx ./seed/newtype/amadeus/container/bin/uvx

mkdir -p ./seed/newtype/amadeus/container/amadeus/
./seed/newtype/amadeus/playpen/container/build.sh
"${container_engine}" save "ghcr.io/ndscm/seed-newtype-amadeus-playpen-container:latest" \
  >"./seed/newtype/amadeus/container/amadeus/playpen.tar"
bazel build --stamp //seed/newtype/amadeus/server
cp -f ./bazel-bin/seed/newtype/amadeus/server/amadeus-server_/amadeus-server ./seed/newtype/amadeus/container/amadeus/amadeus-server

mkdir -p ./seed/newtype/amadeus/container/smd/workstation/
cp -f ./seed/devprod/smd/workstation/setup.sh ./seed/newtype/amadeus/container/smd/workstation/setup.sh

cd ./seed/newtype/amadeus/container/

build_compat=()
if [[ "${container_engine}" == "docker" ]]; then
  build_compat+=("-f" "Containerfile")
elif [[ "${container_engine}" == "podman" ]]; then
  build_compat+=("--userns" "auto:size=65536")
fi

"${container_engine}" build "${build_compat[@]}" -t ghcr.io/ndscm/seed-newtype-amadeus-container:latest .
