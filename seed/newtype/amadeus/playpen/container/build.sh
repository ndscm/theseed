#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

container_engine=${CONTAINER_ENGINE:-"podman"}

export CONTAINER_ENGINE="${container_engine}"
./seed/devprod/container/ubuntu/build.sh

mkdir -p ./seed/newtype/amadeus/playpen/container/bin/
cp -a -f ./seed/vendor/bazel/bin/bazel ./seed/newtype/amadeus/playpen/container/bin/bazel
cp -a -f ./seed/vendor/bazel/bin/bazelisk ./seed/newtype/amadeus/playpen/container/bin/bazelisk
cp -a -f ./seed/vendor/bazel/bin/bazelisk.dotslash ./seed/newtype/amadeus/playpen/container/bin/bazelisk.dotslash
cp -a -f ./seed/vendor/buck/bin/buck2 ./seed/newtype/amadeus/playpen/container/bin/buck2
cp -a -f ./seed/vendor/buck/bin/buck2.dotslash ./seed/newtype/amadeus/playpen/container/bin/buck2.dotslash
cp -a -f ./seed/vendor/buck/bin/reindeer ./seed/newtype/amadeus/playpen/container/bin/reindeer
cp -a -f ./seed/vendor/buck/bin/reindeer.dotslash ./seed/newtype/amadeus/playpen/container/bin/reindeer.dotslash
cp -a -f ./seed/vendor/claude/bin/claude ./seed/newtype/amadeus/playpen/container/bin/claude
cp -a -f ./seed/vendor/claude/bin/claude.dotslash ./seed/newtype/amadeus/playpen/container/bin/claude.dotslash
cp -a -f ./seed/vendor/crane/bin/crane ./seed/newtype/amadeus/playpen/container/bin/crane
cp -a -f ./seed/vendor/crane/bin/crane.dotslash ./seed/newtype/amadeus/playpen/container/bin/crane.dotslash
cp -a -f ./seed/vendor/docker/bin/docker ./seed/newtype/amadeus/playpen/container/bin/docker
cp -a -f ./seed/vendor/docker/bin/docker.dotslash ./seed/newtype/amadeus/playpen/container/bin/docker.dotslash
cp -a -f ./seed/vendor/github/bin/gh ./seed/newtype/amadeus/playpen/container/bin/gh
cp -a -f ./seed/vendor/github/bin/gh.dotslash ./seed/newtype/amadeus/playpen/container/bin/gh.dotslash
cp -a -f ./seed/vendor/jq/bin/jq ./seed/newtype/amadeus/playpen/container/bin/jq
cp -a -f ./seed/vendor/jq/bin/jq.dotslash ./seed/newtype/amadeus/playpen/container/bin/jq.dotslash
cp -a -f ./seed/vendor/lego/bin/lego ./seed/newtype/amadeus/playpen/container/bin/lego
cp -a -f ./seed/vendor/lego/bin/lego.dotslash ./seed/newtype/amadeus/playpen/container/bin/lego.dotslash
cp -a -f ./seed/vendor/llvm/bin/clang ./seed/newtype/amadeus/playpen/container/bin/clang
cp -a -f ./seed/vendor/llvm/bin/clang-format ./seed/newtype/amadeus/playpen/container/bin/clang-format
cp -a -f ./seed/vendor/llvm/bin/clang-format.dotslash ./seed/newtype/amadeus/playpen/container/bin/clang-format.dotslash
cp -a -f ./seed/vendor/llvm/bin/clang.dotslash ./seed/newtype/amadeus/playpen/container/bin/clang.dotslash
cp -a -f ./seed/vendor/llvm/bin/clangd ./seed/newtype/amadeus/playpen/container/bin/clangd
cp -a -f ./seed/vendor/llvm/bin/clangd.dotslash ./seed/newtype/amadeus/playpen/container/bin/clangd.dotslash
cp -a -f ./seed/vendor/ndscm/bin/ndscm ./seed/newtype/amadeus/playpen/container/bin/ndscm
cp -a -f ./seed/vendor/ndscm/bin/ndscm.dotslash ./seed/newtype/amadeus/playpen/container/bin/ndscm.dotslash
cp -a -f ./seed/vendor/node/bin/corepack ./seed/newtype/amadeus/playpen/container/bin/corepack
cp -a -f ./seed/vendor/node/bin/corepack.dotslash ./seed/newtype/amadeus/playpen/container/bin/corepack.dotslash
cp -a -f ./seed/vendor/node/bin/node ./seed/newtype/amadeus/playpen/container/bin/node
cp -a -f ./seed/vendor/node/bin/node.dotslash ./seed/newtype/amadeus/playpen/container/bin/node.dotslash
cp -a -f ./seed/vendor/node/bin/npm ./seed/newtype/amadeus/playpen/container/bin/npm
cp -a -f ./seed/vendor/node/bin/npm.dotslash ./seed/newtype/amadeus/playpen/container/bin/npm.dotslash
cp -a -f ./seed/vendor/node/bin/npx ./seed/newtype/amadeus/playpen/container/bin/npx
cp -a -f ./seed/vendor/node/bin/npx.dotslash ./seed/newtype/amadeus/playpen/container/bin/npx.dotslash
cp -a -f ./seed/vendor/node/bin/pnpm ./seed/newtype/amadeus/playpen/container/bin/pnpm
cp -a -f ./seed/vendor/node/bin/pnpm.shim ./seed/newtype/amadeus/playpen/container/bin/pnpm.shim
cp -a -f ./seed/vendor/node/bin/pnpx ./seed/newtype/amadeus/playpen/container/bin/pnpx
cp -a -f ./seed/vendor/node/bin/pnpx.shim ./seed/newtype/amadeus/playpen/container/bin/pnpx.shim
cp -a -f ./seed/vendor/node/bin/yarn ./seed/newtype/amadeus/playpen/container/bin/yarn
cp -a -f ./seed/vendor/node/bin/yarn.shim ./seed/newtype/amadeus/playpen/container/bin/yarn.shim
cp -a -f ./seed/vendor/node/bin/yarnpkg ./seed/newtype/amadeus/playpen/container/bin/yarnpkg
cp -a -f ./seed/vendor/node/bin/yarnpkg.shim ./seed/newtype/amadeus/playpen/container/bin/yarnpkg.shim
cp -a -f ./seed/vendor/podman/bin/podman-remote ./seed/newtype/amadeus/playpen/container/bin/podman-remote
cp -a -f ./seed/vendor/podman/bin/podman-remote.dotslash ./seed/newtype/amadeus/playpen/container/bin/podman-remote.dotslash
cp -a -f ./seed/vendor/uv/bin/uv ./seed/newtype/amadeus/playpen/container/bin/uv
cp -a -f ./seed/vendor/uv/bin/uv.dotslash ./seed/newtype/amadeus/playpen/container/bin/uv.dotslash
cp -a -f ./seed/vendor/uv/bin/uvx ./seed/newtype/amadeus/playpen/container/bin/uvx
cp -a -f ./seed/vendor/uv/bin/uvx.dotslash ./seed/newtype/amadeus/playpen/container/bin/uvx.dotslash

mkdir -p ./seed/newtype/amadeus/playpen/container/smd/workstation/
cp -a -f ./seed/devprod/smd/workstation/setup.sh ./seed/newtype/amadeus/playpen/container/smd/workstation/setup.sh

cd ./seed/newtype/amadeus/playpen/container/

build_compat=()
if [[ "${container_engine}" == "docker" ]]; then
  build_compat+=("-f" "Containerfile")
elif [[ "${container_engine}" == "podman" ]]; then
  build_compat+=("--userns" "auto:size=65536")
fi

"${container_engine}" build "${build_compat[@]}" -t ghcr.io/ndscm/seed-newtype-amadeus-playpen-container:latest .
