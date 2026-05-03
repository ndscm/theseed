#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

container_cli=${CONTAINER_CLI:-"docker"}

bazel build //seed/cloud/sfe/server
cp -f ./bazel-bin/seed/cloud/sfe/server/sfe-server_/sfe-server ./seed/cloud/sfe/container/sfe-server

cd ./seed/cloud/sfe/container/
"${container_cli}" build -t ghcr.io/ndscm/seed-cloud-sfe-container:latest .
