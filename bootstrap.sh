#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

export SEED_MONOREPO_BOOTSTRAP=1

bazel build \
  --output_groups=+go_generated_srcs \
  --output_groups=+types \
  $(
    bazel query \
      'filter(".*:ent_full_srcs$", //...)' \
      ' + filter(".*(_go_connect_go)$", //...)' \
      ' + filter(".*(_go_proto)$", //...)' \
      ' + filter(".*(_py_grpc)$", //...)' \
      ' + filter(".*(_py_pb2)$", //...)' \
      ' + filter(".*(_ts_proto)$", //...)'
  )

# Monorepo
export ELECTRON_GET_USE_PROXY=1
bazel run @pnpm//:pnpm -- --dir $PWD install
## Build all dependency packages of apps
bazel run @pnpm//:pnpm -- --dir $PWD recursive \
  --filter "@theseed/*-webapp^..." \
  run build
uv sync

find . -type f -path "*/database/build.sh" -exec bash {} \;
find . -type f -path "*/proto/build.sh" -exec bash {} go py \;
