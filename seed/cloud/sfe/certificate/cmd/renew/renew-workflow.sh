#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../../.."

domain=${1:-"workflow.ndscm.biz"}

bazel run //seed/cloud/sfe/certificate/cmd/renew -- \
  --login_openid_discovery_url "https://account.ndscm.com/realms/ndscm/.well-known/openid-configuration" \
  --login_tier "prod" \
  --sfe_certificate_service_server "https://certificate.sfe.ndscm.com" \
  --verbose \
  "${domain}"
