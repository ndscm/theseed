#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../../..

IFS="+" target="${@:1}"

if [[ "${target}" == "" || "${target}" == *"go"* ]]; then
  if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
    bazel build --output_groups=+go_generated_srcs //seed/newtype/hooin/commute/proto:commute_go_proto
  fi
  mkdir -p ./seed/newtype/hooin/commute/proto/commutepb
  cp -f \
    ./bazel-bin/seed/newtype/hooin/commute/proto/commute_go_proto_/github.com/ndscm/theseed/seed/newtype/hooin/commute/proto/commutepb/commute.pb.go \
    ./seed/newtype/hooin/commute/proto/commutepb/commute.pb.go

  if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
    bazel build //seed/newtype/hooin/commute/proto:commute_go_connect_go
  fi
  mkdir -p ./seed/newtype/hooin/commute/proto/commutepbconnect
  cp -f \
    ./bazel-bin/seed/newtype/hooin/commute/proto/commutepbconnect/commute.connect.go \
    ./seed/newtype/hooin/commute/proto/commutepbconnect/commute.connect.go
fi
