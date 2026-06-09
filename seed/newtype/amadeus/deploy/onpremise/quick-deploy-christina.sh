#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

set +x
password="${1:-""}"

./deploy-docker.sh "steins.ndscm.biz" "christina" "2447" "${password}"
