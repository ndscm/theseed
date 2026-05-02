#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

container_cli=${CONTAINER_CLI:-"docker"}

bazel build //seed/office/stuff/server:stuff-server_tar_gz
cp -f ./bazel-bin/seed/office/stuff/server/stuff-server.tar.gz ./seed/office/stuff/container/stuff-server.tar.gz

cd ./seed/office/stuff/container/
"${container_cli}" build -t ghcr.io/ndscm/seed-office-stuff-container:latest .
