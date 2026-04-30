#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

container_cli=${CONTAINER_CLI:-"docker"}

bazel build //seed/devprod/golink/server:golink-server_tar_gz
cp -f ./bazel-bin/seed/devprod/golink/server/golink-server.tar.gz ./seed/devprod/golink/container/golink-server.tar.gz

cd ./seed/devprod/golink/container/
"${container_cli}" build -t ghcr.io/ndscm/seed-devprod-golink-container:latest .
