#!/usr/bin/env bash
# Deploy the ndscm check workflow as a Cloud Build trigger.
#
# Migrated from seed/devprod/workflow/check/Jenkinsfile. Builds and pushes the
# Cloud Build builder image, then creates (or replaces) a GitHub-connected
# trigger on ndscm/theseed that runs cloudbuild.json on push. Cloud Build clones
# the repo at the pushed commit and exposes $COMMIT_SHA / $BRANCH_NAME, which the
# build config forwards to check.sh.
#
# Prerequisites (one-time, not managed here):
#   - The Cloud Build GitHub App is installed on the ndscm/theseed repository.
#   - Secret Manager holds the GitHub status token at
#     projects/ndscm-prod/secrets/ndscm-gh-token, and the Cloud Build service
#     account has roles/secretmanager.secretAccessor on it.
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

container_engine=${CONTAINER_ENGINE:-"podman"}

project="ndscm-prod"
region="us-west1"
image_package="us-docker.pkg.dev/ndscm-prod/container-us/seed-devprod-workflow-builder-deploy-gcloud"

# Build and push the builder image the Cloud Build steps run in.
bazel run --stamp \
  --@rules_img//img/settings:load_daemon="${container_engine}" \
  //seed/devprod/workflow/builder/deploy/gcloud:load-prod
"${container_engine}" push "${image_package}:prod"

image_digest=$(crane digest "${image_package}:prod")
