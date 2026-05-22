#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

bazel run //seed/office/stuff/server -- \
  --stuff_database_secret_file="${ND_USER_SECRET_HOME}/stuff/STUFF_DATABASE_SECRET" \
  --verbose
