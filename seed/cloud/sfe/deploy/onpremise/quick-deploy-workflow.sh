#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

set +x
sfe_openid_client_secret="${1:-""}"

export EXTRA_SFE_ENV="
SFE_WORKFLOW_SERVICE_SERVER=http://127.0.0.1:8080
"

./deploy.sh "jenkins-controller" "sfe" "workflow-sfe-prod" "${sfe_openid_client_secret}"
