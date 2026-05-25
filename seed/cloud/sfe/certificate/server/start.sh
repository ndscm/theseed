#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

bazel run //seed/cloud/sfe/certificate/server -- \
  --acme_provider="letsencrypt-staging" \
  --trust_openid_issuers_file /mnt/data/sfe-certificate/openid_issuers.json \
  --cloudflare_dns_api_token_file="${ND_USER_SECRET_HOME}/sfe-certificate/CLOUDFLARE_DNS_API_TOKEN/latest" \
  --verbose
