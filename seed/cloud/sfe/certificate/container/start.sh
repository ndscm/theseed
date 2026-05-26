#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

container_engine=${CONTAINER_ENGINE:-"podman"}

export CONTAINER_ENGINE="${container_engine}"
./build.sh

# userns keep-id is used to allow the container to read/write files as the current user
"${container_engine}" run --name "sfe-certificate" --rm --interactive --tty \
  --network "host" \
  --userns "keep-id" \
  --volume /mnt/data/sfe-certificate:/mnt/data/sfe-certificate \
  --secret CLOUDFLARE_DNS_API_TOKEN \
  ghcr.io/ndscm/seed-cloud-sfe-certificate-container:latest \
  --acme_provider letsencrypt-staging \
  --trust_openid_issuers_file /mnt/data/sfe-certificate/openid_issuers.json \
  --cloudflare_dns_api_token_file /run/secrets/CLOUDFLARE_DNS_API_TOKEN \
  --verbose
