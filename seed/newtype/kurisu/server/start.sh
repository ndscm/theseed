#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

stack=${1:-"local"}

hooin_dictate_service_server="http://127.0.0.1:4664"
hooin_roster_service_server="http://127.0.0.1:4664"

if [[ "$stack" == "onpremise" ]]; then
  hooin_dictate_service_server="http://steins.ndscm.biz:4664"
  hooin_roster_service_server="http://steins.ndscm.biz:4664"
fi

bazel run //seed/newtype/kurisu/server -- \
  --hooin_dictate_service_server="${hooin_dictate_service_server}" \
  --hooin_roster_service_server="${hooin_roster_service_server}" \
  --verbose
