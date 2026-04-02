#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

IFS="+" target="${@:1}"

if [[ "${target}" == "" || "${target}" == *"go"* ]]; then
  if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
    bazel build --output_groups=+go_generated_srcs //seed/cloud/login/proto:login_go_proto
  fi
  mkdir -p ./seed/cloud/login/proto/loginpb
  cp -f \
    ./bazel-bin/seed/cloud/login/proto/login_go_proto_/github.com/ndscm/theseed/seed/cloud/login/proto/loginpb/login.pb.go \
    ./seed/cloud/login/proto/loginpb/login.pb.go

  if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
    bazel build //seed/cloud/login/proto:login_go_connect_go
  fi
  mkdir -p ./seed/cloud/login/proto/loginpbconnect
  cp -f \
    ./bazel-bin/seed/cloud/login/proto/loginpbconnect/login.connect.go \
    ./seed/cloud/login/proto/loginpbconnect/login.connect.go
fi

if [[ "${target}" == "" || "${target}" == *"ts"* ]]; then
  if [[ -z "${SEED_MONOREPO_BOOTSTRAP:-}" ]]; then
    bazel build --output_groups=+types //seed/cloud/login/proto:login_ts_proto
  fi
  cp -f ./bazel-bin/seed/cloud/login/proto/login_pb.js ./seed/cloud/login/proto/login_pb.js
  cp -f ./bazel-bin/seed/cloud/login/proto/login_pb.d.ts ./seed/cloud/login/proto/login_pb.d.ts
fi
