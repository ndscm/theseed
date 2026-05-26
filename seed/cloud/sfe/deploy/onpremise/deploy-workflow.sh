#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

set +x
sfe_openid_client_secret="${1:-""}"

./deploy.sh "jenkins-controller" "sfe" "workflow-sfe-prod" "${sfe_openid_client_secret}"
