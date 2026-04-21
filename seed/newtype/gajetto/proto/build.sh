#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

IFS="+" target="${@:1}"

if [[ "${target}" == "" || "${target}" == *"go"* ]]; then
  if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
    bazel build --output_groups=+go_generated_srcs //seed/newtype/gajetto/proto:brain_go_proto
  fi
  mkdir -p ./seed/newtype/gajetto/proto/brainpb
  cp -f \
    ./bazel-bin/seed/newtype/gajetto/proto/brain_go_proto_/github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb/brain.pb.go \
    ./seed/newtype/gajetto/proto/brainpb/brain.pb.go
fi
