#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

jenkins_user=${1:-""}
jenkins_api_token=${2:-""}

./deploy-relay.sh \
  "workflow.ndscm.biz" \
  "jenkins-controller" \
  "https://webhook.ndscm.com/ndscm/github/subscribe" \
  "http://127.0.0.1:8080/generic-webhook-trigger/invoke" \
  "${jenkins_user}" \
  "${jenkins_api_token}"
