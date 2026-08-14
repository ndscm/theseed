#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

# It's recommended to use on-premise RBE proxy for Bazel remote cache,
# which is faster and more reliable than direct GCS access.
target="${1:-"http://cache.rbe.ndscm.biz:22243"}"

cat >rbe.local.bazelrc <<EOF
common --remote_cache=${target}
EOF
