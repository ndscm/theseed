#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

container_engine=${CONTAINER_ENGINE:-"podman"}
if [[ "${container_engine}" != "podman" ]]; then
  printf 'Only podman is supported for now\n'
  exit 1
fi

bazel build //seed/devprod/ndscm/cli
cp -f ./bazel-bin/seed/devprod/ndscm/cli/ndscm_/ndscm ./seed/devprod/jenkins/node/container/ndscm

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
