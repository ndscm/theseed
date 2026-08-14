#!/usr/bin/env bash
set -euo pipefail

# Bazel credential helper for the GCS remote cache.
#
# Not `--google_default_credentials`: Google Cloud requires a RAPT (reauth) on
# every token refresh. Bazel's Java auth library can't do the reauth flow (fails
# with `invalid_rapt`); gcloud can, so we let it mint the token here. Protocol:
# Bazel writes a request on stdin, reads the response on stdout.

token_full="$(gcloud auth application-default print-access-token --format json)"

access_token="$(jq -r '.token' <<<"${token_full}")"
expiry_datetime="$(jq -r '.expiry.datetime' <<<"${token_full}")"
expires="$(date -u -d "${expiry_datetime} UTC" +%Y-%m-%dT%H:%M:%SZ)"

printf '{"headers":{"Authorization":["Bearer %s"]},"expires":"%s"}' "${access_token}" "${expires}"
