#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

export MOUNT_JENKINS_HOME="/mnt/data/workflow/jenkins_home"

./deploy.sh "workflow.ndscm.biz" "jenkins-controller" "8080"
