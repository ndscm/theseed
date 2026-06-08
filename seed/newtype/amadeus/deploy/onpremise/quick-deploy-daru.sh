#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

set +x
refresh_token="${1:-""}"

./deploy.sh \
  "steins.ndscm.biz" \
  "daru" \
  "3278" \
  "https://hooin.ndscm.biz/" \
  "daru" \
  "${refresh_token}"
