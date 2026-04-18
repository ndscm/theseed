#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../../..

IFS="+" target="${@:1}"

if [[ "${target}" == "" || "${target}" == *"go"* ]]; then
  if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
    bazel build --output_groups=+go_generated_srcs //seed/newtype/amadeus/wake/proto:wake_go_proto
  fi
  mkdir -p ./seed/newtype/amadeus/wake/proto/wakepb
  cp -f \
    ./bazel-bin/seed/newtype/amadeus/wake/proto/wake_go_proto_/github.com/ndscm/theseed/seed/newtype/amadeus/wake/proto/wakepb/wake.pb.go \
    ./seed/newtype/amadeus/wake/proto/wakepb/wake.pb.go

  if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
    bazel build //seed/newtype/amadeus/wake/proto:wake_go_connect_go
  fi
  mkdir -p ./seed/newtype/amadeus/wake/proto/wakepbconnect
  cp -f \
    ./bazel-bin/seed/newtype/amadeus/wake/proto/wakepbconnect/wake.connect.go \
    ./seed/newtype/amadeus/wake/proto/wakepbconnect/wake.connect.go
fi
