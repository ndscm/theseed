#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

person=${1:-"christina"}
topic=${2:-"review"}
text=${3:-"I need to review the pull requests."}

bazel run //seed/newtype/hooin/cmd/send -- \
  --hooin_dictate_service_server http://newtype.ndscm.com:4664 \
  --login_tier prod \
  --person "${person}" \
  --topic "${topic}" \
  --stream \
  --text "${text}" \
  --verbose
