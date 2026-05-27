#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

container_engine=${CONTAINER_ENGINE:-"podman"}

project="ndscm-prod"
region="us-west1"
service="seed-cloud-sfe-prod"
image_package="us-docker.pkg.dev/ndscm-prod/container-us/seed-cloud-sfe-deploy-gcloud"

export CONTAINER_ENGINE="${container_engine}"
./seed/cloud/sfe/container/build.sh

cd ./seed/cloud/sfe/deploy/gcloud/

build_compat=()
if [[ "${container_engine}" == "docker" ]]; then
  build_compat+=("-f" "Containerfile")
elif [[ "${container_engine}" == "podman" ]]; then
  build_compat+=("--userns" "auto:size=65536")
fi

"${container_engine}" build "${build_compat[@]}" -t "${image_package}:prod" .

"${container_engine}" push "${image_package}:prod"

image_digest=$(crane digest "${image_package}:prod")
gcloud run services update "${service}" --project="${project}" --region="${region}" --image="${image_package}@${image_digest}"
