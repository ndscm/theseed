#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

bazel build //seed/devprod/webhook/relay
cp -f ./bazel-bin/seed/devprod/webhook/relay/webhook-relay_/webhook-relay ./seed/devprod/webhook/relay/container/webhook-relay

cd ./seed/devprod/webhook/relay/container/
"${CONTAINER_CLI}" build -t ghcr.io/ndscm/seed-devprod-webhook-relay-container:latest .
