#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

container_engine=${CONTAINER_ENGINE:-"podman"}

bazel build //seed/devprod/container/ubuntu:ubuntu_2604_freedom

"${container_engine}" load \
  -i "bazel-bin/seed/devprod/container/ubuntu/ubuntu_2604_freedom.tar"
"${container_engine}" tag \
  ghcr.io/ndscm/ubuntu:2604-freedom \
  ghcr.io/ndscm/ubuntu:latest
