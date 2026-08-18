#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

version="${1:-"581.0.0"}"

# See: https://console.cloud.google.com/storage/browser/cloud-sdk-release

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/gcloud/gcloud.dotslash.json" \
  --replace "VERSION=${version}" \
  --outdir "$(pwd)/seed/vendor/gcloud/bin"

chmod +x ./seed/vendor/gcloud/bin/gcloud.dotslash

ln -s -f gcloud.dotslash ./seed/vendor/gcloud/bin/gcloud
