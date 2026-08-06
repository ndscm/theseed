#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

container_engine=${CONTAINER_ENGINE:-"podman"}

project="ndscm-prod"
region="us-west1"
service="seed-newtype-kurisu-prod"
image_package="us-docker.pkg.dev/ndscm-prod/container-us/seed-newtype-kurisu-deploy-gcloud"

bazel run --stamp \
  --@rules_img//img/settings:load_daemon="${container_engine}" \
  //seed/newtype/kurisu/deploy/gcloud:load-prod
"${container_engine}" push "${image_package}:prod"

image_digest=$(crane digest "${image_package}:prod")
gcloud run services update \
  --project="${project}" \
  --region="${region}" \
  "${service}" \
  --image="${image_package}@${image_digest}"
