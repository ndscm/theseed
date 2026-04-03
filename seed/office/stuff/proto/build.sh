#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

IFS="+" target="${@:1}"

if [[ "${target}" == "" || "${target}" == *"go"* ]]; then
  if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
    bazel build --output_groups=+go_generated_srcs //seed/office/stuff/proto:stuff_go_proto
  fi
  mkdir -p ./seed/office/stuff/proto/stuffpb
  cp -f \
    ./bazel-bin/seed/office/stuff/proto/stuff_go_proto_/github.com/ndscm/theseed/seed/office/stuff/proto/stuffpb/stuff.pb.go \
    ./seed/office/stuff/proto/stuffpb/stuff.pb.go

  if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
    bazel build //seed/office/stuff/proto:stuff_go_connect_go
  fi
  mkdir -p ./seed/office/stuff/proto/stuffpbconnect
  cp -f \
    ./bazel-bin/seed/office/stuff/proto/stuffpbconnect/stuff.connect.go \
    ./seed/office/stuff/proto/stuffpbconnect/stuff.connect.go

  if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
    bazel build --output_groups=+go_generated_srcs //seed/office/stuff/proto:stuff_error_go_proto
  fi
  mkdir -p ./seed/office/stuff/proto/stufferrorpb
  cp -f \
    ./bazel-bin/seed/office/stuff/proto/stuff_error_go_proto_/github.com/ndscm/theseed/seed/office/stuff/proto/stufferrorpb/stuff_error.pb.go \
    ./seed/office/stuff/proto/stufferrorpb/stuff_error.pb.go
fi

if [[ "${target}" == "" || "${target}" == *"ts"* ]]; then
  if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
    bazel build --output_groups=+types //seed/office/stuff/proto:stuff_ts_proto
  fi
  cp -f ./bazel-bin/seed/office/stuff/proto/stuff_pb.js ./seed/office/stuff/proto/stuff_pb.js
  cp -f ./bazel-bin/seed/office/stuff/proto/stuff_pb.d.ts ./seed/office/stuff/proto/stuff_pb.d.ts
  if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
    bazel build --output_groups=+types //seed/office/stuff/proto:stuff_error_ts_proto
  fi
  cp -f ./bazel-bin/seed/office/stuff/proto/stuff_error_pb.js ./seed/office/stuff/proto/stuff_error_pb.js
  cp -f ./bazel-bin/seed/office/stuff/proto/stuff_error_pb.d.ts ./seed/office/stuff/proto/stuff_error_pb.d.ts
fi
