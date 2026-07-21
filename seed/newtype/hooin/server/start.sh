#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

bazel run //seed/newtype/hooin/server -- \
  --openid_client_id "hooin-dev-${ND_USER_HANDLE}" \
  --openid_client_secret_file "sops:$(ndscm secret --user get-path seed/newtype/hooin/server/OPENID_CLIENT_SECRET.sops)" \
  --verbose
