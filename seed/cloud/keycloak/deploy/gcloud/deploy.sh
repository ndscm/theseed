#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

container_cli=${CONTAINER_CLI:-"docker"}

project="ndscm-prod"
region="us-west1"
service="seed-cloud-keycloak-prod"
image_package="us-docker.pkg.dev/ndscm-prod/container-us/seed-cloud-keycloak-deploy-gcloud"

./seed/cloud/keycloak/container/build.sh

cd ./seed/cloud/keycloak/deploy/gcloud/
"${container_cli}" build -t "${image_package}:prod" .

"${container_cli}" push "${image_package}:prod"

image_digest="$("${container_cli}" inspect --format='{{index .RepoDigests 0}}' "${image_package}:prod")"
gcloud run services update "${service}" \
  --project="${project}" \
  --region="${region}" \
  --image="${image_digest}" \
  --startup-probe="httpGet.port=9000,httpGet.path=/health/ready,initialDelaySeconds=1,periodSeconds=1,timeoutSeconds=1,failureThreshold=120"
