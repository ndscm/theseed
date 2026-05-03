#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

project=${1:-"ndscm-prod"}
service_account=${2:-"keycloak-prod@$project.iam.gserviceaccount.com"}

gcloud projects add-iam-policy-binding "${project}" \
  --member="serviceAccount:${service_account}" \
  --role="roles/cloudsql.client"
