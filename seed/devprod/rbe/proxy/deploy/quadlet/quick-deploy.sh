#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

set +x
openid_client_secret="${1:-""}"

./deploy.sh \
  "cache.rbe.ndscm.biz" \
  "rbe" \
  "22243" \
  "https://ndscm-prod-rbe-cache-prod.storage.googleapis.com/" \
  "${openid_client_secret}"
