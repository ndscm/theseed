#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

container_engine=${CONTAINER_ENGINE:-"podman"}

export CONTAINER_ENGINE="${container_engine}"
./seed/devprod/container/ubuntu/build.sh

mkdir -p ./seed/newtype/amadeus/playpen/container/bin/

xslash -- create --skeleton ./seed/prebuilt/age/bin/age-inspect.xslash --out ./seed/newtype/amadeus/playpen/container/bin/age-inspect.xslash --symlink age-inspect
xslash -- create --skeleton ./seed/prebuilt/age/bin/age-keygen.xslash --out ./seed/newtype/amadeus/playpen/container/bin/age-keygen.xslash --symlink age-keygen
xslash -- create --skeleton ./seed/prebuilt/age/bin/age.xslash --out ./seed/newtype/amadeus/playpen/container/bin/age.xslash --symlink age
xslash -- create --skeleton ./seed/prebuilt/github/bin/gh.xslash --out ./seed/newtype/amadeus/playpen/container/bin/gh.xslash --symlink gh
xslash -- create --skeleton ./seed/prebuilt/lego/bin/lego.xslash --out ./seed/newtype/amadeus/playpen/container/bin/lego.xslash --symlink lego
xslash -- create --skeleton ./seed/prebuilt/ndscm/bin/ndscm.xslash --out ./seed/newtype/amadeus/playpen/container/bin/ndscm.xslash --symlink nd --symlink ndscm
xslash -- create --skeleton ./seed/prebuilt/node/bin/node.xslash --out ./seed/newtype/amadeus/playpen/container/bin/node.xslash --symlink node
xslash -- create --skeleton ./seed/prebuilt/node/bin/npm.xslash --out ./seed/newtype/amadeus/playpen/container/bin/npm.xslash --symlink npm
xslash -- create --skeleton ./seed/prebuilt/node/bin/npx.xslash --out ./seed/newtype/amadeus/playpen/container/bin/npx.xslash --symlink npx
xslash -- create --skeleton ./seed/prebuilt/pnpm/bin/pnpm.xslash --out ./seed/newtype/amadeus/playpen/container/bin/pnpm.xslash --symlink pnpm
xslash -- create --skeleton ./seed/prebuilt/python/bin/python3.xslash --out ./seed/newtype/amadeus/playpen/container/bin/python3.xslash --symlink python --symlink python3
xslash -- create --skeleton ./seed/prebuilt/shfmt/bin/shfmt.xslash --out ./seed/newtype/amadeus/playpen/container/bin/shfmt.xslash --symlink shfmt
xslash -- create --skeleton ./seed/prebuilt/sops/bin/sops.xslash --out ./seed/newtype/amadeus/playpen/container/bin/sops.xslash --symlink sops
xslash -- create --skeleton ./seed/prebuilt/uv/bin/uv.xslash --out ./seed/newtype/amadeus/playpen/container/bin/uv.xslash --symlink uv
xslash -- create --skeleton ./seed/prebuilt/uv/bin/uvx.xslash --out ./seed/newtype/amadeus/playpen/container/bin/uvx.xslash --symlink uvx

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
cp -a -f ./seed/vendor/jq/bin/jq ./seed/newtype/amadeus/playpen/container/bin/jq
cp -a -f ./seed/vendor/jq/bin/jq.dotslash ./seed/newtype/amadeus/playpen/container/bin/jq.dotslash
cp -a -f ./seed/vendor/llvm/bin/clang ./seed/newtype/amadeus/playpen/container/bin/clang
cp -a -f ./seed/vendor/llvm/bin/clang-format ./seed/newtype/amadeus/playpen/container/bin/clang-format
cp -a -f ./seed/vendor/llvm/bin/clang-format.dotslash ./seed/newtype/amadeus/playpen/container/bin/clang-format.dotslash
cp -a -f ./seed/vendor/llvm/bin/clang.dotslash ./seed/newtype/amadeus/playpen/container/bin/clang.dotslash
cp -a -f ./seed/vendor/llvm/bin/clangd ./seed/newtype/amadeus/playpen/container/bin/clangd
cp -a -f ./seed/vendor/llvm/bin/clangd.dotslash ./seed/newtype/amadeus/playpen/container/bin/clangd.dotslash
cp -a -f ./seed/vendor/podman/bin/podman-remote ./seed/newtype/amadeus/playpen/container/bin/podman-remote
cp -a -f ./seed/vendor/podman/bin/podman-remote.dotslash ./seed/newtype/amadeus/playpen/container/bin/podman-remote.dotslash

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
