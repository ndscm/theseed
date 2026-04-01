#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

IFS="+" target="${@:1}"

if [[ "${target}" == "" || "${target}" == *"go"* ]]; then
  bazel build --output_groups=+go_generated_srcs //seed/devprod/golink/proto:golink_go_proto
  mkdir -p ./seed/devprod/golink/proto/golinkpb
  cp -f \
    ./bazel-bin/seed/devprod/golink/proto/golink_go_proto_/github.com/ndscm/theseed/seed/devprod/golink/proto/golinkpb/golink.pb.go \
    ./seed/devprod/golink/proto/golinkpb/golink.pb.go

  bazel build //seed/devprod/golink/proto:golink_go_connect_go
  mkdir -p ./seed/devprod/golink/proto/golinkpbconnect
  cp -f \
    ./bazel-bin/seed/devprod/golink/proto/golinkpbconnect/golink.connect.go \
    ./seed/devprod/golink/proto/golinkpbconnect/golink.connect.go

  bazel build --output_groups=+go_generated_srcs //seed/devprod/golink/proto:golink_error_go_proto
  mkdir -p ./seed/devprod/golink/proto/golinkerrorpb
  cp -f \
    ./bazel-bin/seed/devprod/golink/proto/golink_error_go_proto_/github.com/ndscm/theseed/seed/devprod/golink/proto/golinkerrorpb/golink_error.pb.go \
    ./seed/devprod/golink/proto/golinkerrorpb/golink_error.pb.go
fi

if [[ "${target}" == "" || "${target}" == *"ts"* ]]; then
  bazel build --output_groups=+types //seed/devprod/golink/proto:golink_ts_proto
  cp -f ./bazel-bin/seed/devprod/golink/proto/golink_pb.js ./seed/devprod/golink/proto/golink_pb.js
  cp -f ./bazel-bin/seed/devprod/golink/proto/golink_pb.d.ts ./seed/devprod/golink/proto/golink_pb.d.ts
  bazel build --output_groups=+types //seed/devprod/golink/proto:golink_error_ts_proto
  cp -f ./bazel-bin/seed/devprod/golink/proto/golink_error_pb.js ./seed/devprod/golink/proto/golink_error_pb.js
  cp -f ./bazel-bin/seed/devprod/golink/proto/golink_error_pb.d.ts ./seed/devprod/golink/proto/golink_error_pb.d.ts
fi
