#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

"${CONTAINER_CLI}" build -t ghcr.io/ndscm/seed-devprod-rbe-cache:latest .
