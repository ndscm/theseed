#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../../..

IFS="+" target="${@:1}"

if [[ "${target}" == "" || "${target}" == *"go"* ]]; then
  if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
    bazel build --output_groups=+go_generated_srcs //seed/newtype/hooin/dictate/proto:dictate_go_proto
  fi
  mkdir -p ./seed/newtype/hooin/dictate/proto/dictatepb
  cp -f \
    ./bazel-bin/seed/newtype/hooin/dictate/proto/dictate_go_proto_/github.com/ndscm/theseed/seed/newtype/hooin/dictate/proto/dictatepb/dictate.pb.go \
    ./seed/newtype/hooin/dictate/proto/dictatepb/dictate.pb.go

  if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
    bazel build //seed/newtype/hooin/dictate/proto:dictate_go_connect_go
  fi
  mkdir -p ./seed/newtype/hooin/dictate/proto/dictatepbconnect
  cp -f \
    ./bazel-bin/seed/newtype/hooin/dictate/proto/dictatepbconnect/dictate.connect.go \
    ./seed/newtype/hooin/dictate/proto/dictatepbconnect/dictate.connect.go
fi
