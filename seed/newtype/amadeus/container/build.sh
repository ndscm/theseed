#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

bazel build //seed/newtype/amadeus/server
cp -f ./bazel-bin/seed/newtype/amadeus/server/amadeus-server_/amadeus-server ./seed/newtype/amadeus/container/amadeus-server

cd ./seed/newtype/amadeus/container/
"${CONTAINER_CLI}" build -t ghcr.io/ndscm/seed-newtype-amadeus-container:latest .
