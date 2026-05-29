#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

token=${1:-"christina:unsecure"}

bazel run //seed/newtype/amadeus/cmd/wake -- \
  --amadeus_service_server http://newtype.ndscm.com:2447 \
  --login_tier prod \
  --hooin_direct_server http://newtype.ndscm.com:4664 \
  --token "${token}" \
  --verbose
