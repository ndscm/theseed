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
xslash -- create --skeleton ./seed/prebuilt/atlas/bin/atlas.xslash --out ./seed/newtype/amadeus/playpen/container/bin/atlas.xslash --symlink atlas
xslash -- create --skeleton ./seed/prebuilt/bazel/bin/bazelisk.xslash --out ./seed/newtype/amadeus/playpen/container/bin/bazelisk.xslash --symlink bazel --symlink bazelisk
xslash -- create --skeleton ./seed/prebuilt/buck/bin/buck2.xslash --out ./seed/newtype/amadeus/playpen/container/bin/buck2.xslash --symlink buck2
xslash -- create --skeleton ./seed/prebuilt/crane/bin/crane.xslash --out ./seed/newtype/amadeus/playpen/container/bin/crane.xslash --symlink crane
xslash -- create --skeleton ./seed/prebuilt/docker/bin/docker.xslash --out ./seed/newtype/amadeus/playpen/container/bin/docker.xslash --symlink docker
xslash -- create --skeleton ./seed/prebuilt/gcloud/bin/gcloud.xslash --out ./seed/newtype/amadeus/playpen/container/bin/gcloud.xslash --symlink gcloud
xslash -- create --skeleton ./seed/prebuilt/github/bin/gh.xslash --out ./seed/newtype/amadeus/playpen/container/bin/gh.xslash --symlink gh
xslash -- create --skeleton ./seed/prebuilt/jq/bin/jq.xslash --out ./seed/newtype/amadeus/playpen/container/bin/jq.xslash --symlink jq
xslash -- create --skeleton ./seed/prebuilt/kubectl/bin/kubectl.xslash --out ./seed/newtype/amadeus/playpen/container/bin/kubectl.xslash --symlink kubectl
xslash -- create --skeleton ./seed/prebuilt/lego/bin/lego.xslash --out ./seed/newtype/amadeus/playpen/container/bin/lego.xslash --symlink lego
xslash -- create --skeleton ./seed/prebuilt/minikube/bin/minikube.xslash --out ./seed/newtype/amadeus/playpen/container/bin/minikube.xslash --symlink minikube
xslash -- create --skeleton ./seed/prebuilt/ndscm/bin/ndscm.xslash --out ./seed/newtype/amadeus/playpen/container/bin/ndscm.xslash --symlink nd --symlink ndscm
xslash -- create --skeleton ./seed/prebuilt/node/bin/node.xslash --out ./seed/newtype/amadeus/playpen/container/bin/node.xslash --symlink node
xslash -- create --skeleton ./seed/prebuilt/node/bin/npm.xslash --out ./seed/newtype/amadeus/playpen/container/bin/npm.xslash --symlink npm
xslash -- create --skeleton ./seed/prebuilt/node/bin/npx.xslash --out ./seed/newtype/amadeus/playpen/container/bin/npx.xslash --symlink npx
xslash -- create --skeleton ./seed/prebuilt/pnpm/bin/pnpm.xslash --out ./seed/newtype/amadeus/playpen/container/bin/pnpm.xslash --symlink pnpm
xslash -- create --skeleton ./seed/prebuilt/podman/bin/podman-remote.xslash --out ./seed/newtype/amadeus/playpen/container/bin/podman-remote.xslash --symlink podman-remote
xslash -- create --skeleton ./seed/prebuilt/python/bin/python3.xslash --out ./seed/newtype/amadeus/playpen/container/bin/python3.xslash --symlink python --symlink python3
xslash -- create --skeleton ./seed/prebuilt/reindeer/bin/reindeer.xslash --out ./seed/newtype/amadeus/playpen/container/bin/reindeer.xslash --symlink reindeer
xslash -- create --skeleton ./seed/prebuilt/shfmt/bin/shfmt.xslash --out ./seed/newtype/amadeus/playpen/container/bin/shfmt.xslash --symlink shfmt
xslash -- create --skeleton ./seed/prebuilt/sops/bin/sops.xslash --out ./seed/newtype/amadeus/playpen/container/bin/sops.xslash --symlink sops
xslash -- create --skeleton ./seed/prebuilt/uv/bin/uv.xslash --out ./seed/newtype/amadeus/playpen/container/bin/uv.xslash --symlink uv
xslash -- create --skeleton ./seed/prebuilt/uv/bin/uvx.xslash --out ./seed/newtype/amadeus/playpen/container/bin/uvx.xslash --symlink uvx

cp -a -f ./seed/vendor/claude/bin/claude ./seed/newtype/amadeus/playpen/container/bin/claude
cp -a -f ./seed/vendor/claude/bin/claude.dotslash ./seed/newtype/amadeus/playpen/container/bin/claude.dotslash
cp -a -f ./seed/vendor/llvm/bin/clang ./seed/newtype/amadeus/playpen/container/bin/clang
cp -a -f ./seed/vendor/llvm/bin/clang-format ./seed/newtype/amadeus/playpen/container/bin/clang-format
cp -a -f ./seed/vendor/llvm/bin/clang-format.dotslash ./seed/newtype/amadeus/playpen/container/bin/clang-format.dotslash
cp -a -f ./seed/vendor/llvm/bin/clang.dotslash ./seed/newtype/amadeus/playpen/container/bin/clang.dotslash
cp -a -f ./seed/vendor/llvm/bin/clangd ./seed/newtype/amadeus/playpen/container/bin/clangd
cp -a -f ./seed/vendor/llvm/bin/clangd.dotslash ./seed/newtype/amadeus/playpen/container/bin/clangd.dotslash

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
