#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

container_engine=${CONTAINER_ENGINE:-"podman"}
if [[ "${container_engine}" != "podman" ]]; then
  printf 'Only podman is supported for now\n'
  exit 1
fi

export CONTAINER_ENGINE="${container_engine}"
./seed/devprod/container/debian/build.sh

mkdir -p ./seed/devprod/jenkins/node/container/bin/

cp -f ./seed/vendor/docker/bin/docker ./seed/devprod/jenkins/node/container/bin/docker

xslash -- create --skeleton ./seed/prebuilt/github/bin/gh.xslash --out ./seed/devprod/jenkins/node/container/bin/gh.xslash --symlink gh
xslash -- create --skeleton ./seed/prebuilt/go/bin/go.xslash --out ./seed/devprod/jenkins/node/container/bin/go.xslash --symlink go
xslash -- create --skeleton ./seed/prebuilt/go/bin/gofmt.xslash --out ./seed/devprod/jenkins/node/container/bin/gofmt.xslash --symlink gofmt
xslash -- create --skeleton ./seed/prebuilt/ndscm/bin/ndscm.xslash --out ./seed/devprod/jenkins/node/container/bin/ndscm.xslash --symlink nd --symlink ndscm

cd ./seed/devprod/jenkins/node/container/

# --userns=auto: required for rootless podman to correctly preserve file
#   ownership in the built image. Without it, podman mangles uid 1000
#   (jenkins) to uid 0 (root) in the stored layers.
#
# size=65536: ensures the nobody user (uid 6553x) falls within the
#   allocated subordinate UID range.
podman build \
  --userns=auto:size=65536 \
  -t ghcr.io/ndscm/seed-devprod-jenkins-node-container:latest \
  .
