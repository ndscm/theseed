#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

set +x
refresh_token="${1:-""}"

./deploy.sh \
  "steins.ndscm.biz" \
  "okarin" \
  "6527" \
  "https://hooin.ndscm.biz/" \
  "okarin" \
  "${refresh_token}"
