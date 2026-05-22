#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

bazel run //seed/devprod/golink/server -- \
  --golink_database_secret_file="${ND_USER_SECRET_HOME}/golink/GOLINK_DATABASE_SECRET" \
  --verbose
