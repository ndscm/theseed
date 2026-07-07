#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

set +x
refresh_token="${1:-""}"

./deploy.sh \
  "steins.ndscm.biz" \
  "christina" \
  "2447" \
  "https://hooin.ndscm.biz/" \
  "${refresh_token}"
