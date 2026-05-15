#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

container_engine=${CONTAINER_ENGINE:-"podman"}

"${container_engine}" build -f ./Containerfile -t ghcr.io/ndscm/ubuntu:latest .
