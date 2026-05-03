#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

container_cli=${CONTAINER_CLI:-"docker"}

project="ndscm-prod"
region="us-west1"
service="seed-office-stuff-prod"
image_package="us-docker.pkg.dev/ndscm-prod/container-us/seed-office-stuff-deploy-gcloud"

./seed/office/stuff/container/build.sh

cd ./seed/office/stuff/deploy/gcloud/
"${container_cli}" build -t "${image_package}:prod" .

"${container_cli}" push "${image_package}:prod"

image_digest="$("${container_cli}" inspect --format='{{index .RepoDigests 0}}' "${image_package}:prod")"
gcloud run services update "${service}" --project="${project}" --region="${region}" --image="${image_digest}"
