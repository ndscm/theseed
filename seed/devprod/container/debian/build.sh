#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

container_engine=${CONTAINER_ENGINE:-"podman"}

bazel build //seed/devprod/container/debian:debian_13_freedom

"${container_engine}" load \
  -i "bazel-bin/seed/devprod/container/debian/debian_13_freedom.tar"
"${container_engine}" tag \
  ghcr.io/ndscm/debian:13-freedom \
  ghcr.io/ndscm/debian:latest
