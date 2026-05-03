#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

bazel build //seed/newtype/hooin/server
cp -f ./bazel-bin/seed/newtype/hooin/server/hooin-server_/hooin-server ./seed/newtype/hooin/container/hooin-server

cd ./seed/newtype/hooin/container/
"${CONTAINER_CLI}" build -t ghcr.io/ndscm/seed-newtype-hooin-container:latest .
