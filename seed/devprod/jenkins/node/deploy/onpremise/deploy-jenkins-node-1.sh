#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

set +x
jenkins_agent_secret="${1:-""}"

./deploy.sh "jenkins-node-1" "jenkins-node" "https://workflow.ndscm.com/" "jenkins-node-1" "${jenkins_agent_secret}"
