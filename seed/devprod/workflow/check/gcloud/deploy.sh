#!/usr/bin/env bash
# Deploy the ndscm check workflow as a Cloud Build trigger.
#
# Migrated from seed/devprod/workflow/check/Jenkinsfile. Creates (or replaces) a
# GitHub-connected trigger on ndscm/theseed that runs cloudbuild.json on push.
# Cloud Build clones the repo at the pushed commit and exposes $COMMIT_SHA /
# $BRANCH_NAME, which the build config forwards to check.sh.
#
# The builder image the steps run in is built and pushed by
# seed/devprod/workflow/builder/deploy/gcloud, which this script invokes; the
# trigger is then pinned to the pushed image by digest.
#
# Prerequisites (one-time, not managed here):
#   - The Cloud Build GitHub App is installed on the ndscm/theseed repository.
#   - report.sh posts commit statuses with the github connection's OAuth token.
#     deploy.sh resolves that secret from the connection (so cloudbuild.json need
#     not hardcode its connection-specific name) and passes it as _GH_OAUTH_SECRET;
#     the Cloud Build service account needs roles/secretmanager.secretAccessor on it.
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

project="ndscm-prod"
region="us-west1"
connection="ndscm"
trigger="seed-devprod-workflow-check"
build_config="seed/devprod/workflow/check/gcloud/cloudbuild.json"
builder_image_package="us-docker.pkg.dev/ndscm-prod/container-us/seed-devprod-workflow-builder-deploy-gcloud"

# ndscm-prod has no legacy default Cloud Build service account, so the trigger
# must run as an explicit service account. cloudbuild.json sets
# logging=CLOUD_LOGGING_ONLY, which a user-specified account requires.
service_account="projects/${project}/serviceAccounts/cloudbuild@${project}.iam.gserviceaccount.com"

# Build and push the builder image the Cloud Build steps run in.
../../builder/deploy/gcloud/deploy.sh

builder_image_digest=$(crane digest "${builder_image_package}:prod")

# Resolve the connection's OAuth token secret so cloudbuild.json need not hardcode
# its connection-specific name; report.sh reads it to post commit statuses.
oauth_token_secret=$(gcloud builds connections describe "${connection}" \
  --project="${project}" \
  --region="${region}" \
  --format="value(githubConfig.authorizerCredential.oauthTokenSecretVersion)")

# The github connection and its repository are 2nd-gen (regional) resources, so
# the trigger must live in the same region; describe/delete/create all need it.
#
# Recreate the trigger so repeated deploys stay idempotent; github triggers have
# no in-place flag update.
if gcloud builds triggers describe "${trigger}" --project="${project}" --region="${region}" >/dev/null 2>&1; then
  gcloud builds triggers delete "${trigger}" --project="${project}" --region="${region}" --quiet
fi

gcloud builds triggers create github \
  --project="${project}" \
  --region="${region}" \
  --name="${trigger}" \
  --repository="projects/${project}/locations/${region}/connections/${connection}/repositories/ndscm-theseed" \
  --branch-pattern=".*" \
  --build-config="${build_config}" \
  --service-account="${service_account}" \
  --substitutions="_BUILDER_IMAGE=${builder_image_package}@${builder_image_digest},_GH_OAUTH_SECRET=${oauth_token_secret}"
