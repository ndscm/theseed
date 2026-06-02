#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

set +x
sfe_openid_client_secret="${1:-""}"

export EXTRA_SFE_CONTAINER_CONFIG="
Secret=KURISU_OPENID_CLIENT_SECRET
"

export EXTRA_SFE_ENV="
SFE_KURISU_OPENID_DISCOVERY_URL=https://account.ndscm.com/realms/ndscm/.well-known/openid-configuration
SFE_KURISU_OPENID_CLIENT_ID=kurisu-prod
SFE_KURISU_OPENID_CLIENT_SECRET_FILE=/run/secrets/KURISU_OPENID_CLIENT_SECRET
SFE_KURISU_SERVICE_SERVER=http://127.0.0.1:5874
"
./deploy.sh "kurisu.ndscm.biz" "sfe" "kurisu-sfe-prod" "${sfe_openid_client_secret}"
