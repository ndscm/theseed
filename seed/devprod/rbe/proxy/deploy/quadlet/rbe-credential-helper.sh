#!/usr/bin/env bash
set -eu
set -o pipefail

export GOOGLE_EXTERNAL_ACCOUNT_ALLOW_EXECUTABLES="1"
export GOOGLE_APPLICATION_CREDENTIALS="/opt/rbe/gcloud-credentials.json"
/opt/gcloud/gcloud-credential-helper --output_format "oauth2"
