#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

bazel build //seed/devprod/golink/database:ent_full_srcs
mkdir -p ./seed/devprod/golink/database/ent
rsync --archive --delete \
  --include='*/' --include='*.go' --exclude='*' \
  ./bazel-bin/seed/devprod/golink/database/ent/. \
  ./seed/devprod/golink/database/ent/.
