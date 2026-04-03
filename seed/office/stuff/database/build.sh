#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
  bazel build //seed/office/stuff/database:ent_full_srcs
fi
mkdir -p ./seed/office/stuff/database/ent
rsync --archive --delete \
  --include='*/' --include='*.go' --exclude='*' \
  ./bazel-bin/seed/office/stuff/database/ent/. \
  ./seed/office/stuff/database/ent/.
