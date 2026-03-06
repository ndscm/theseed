#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

# Monorepo
export ELECTRON_GET_USE_PROXY=1
bazel run @pnpm//:pnpm -- --dir $PWD install
## Build all dependency packages of apps
bazel run @pnpm//:pnpm -- --dir $PWD recursive \
  --filter "@theseed/*proto..." \
  run build
bazel run @pnpm//:pnpm -- --dir $PWD recursive \
  --filter "@theseed/*-ts-service-context..." \
  --filter "@theseed/*haraka..." \
  --filter "@theseed/devprod-buildinfo*..." \
  --filter "@theseed/devprod-gotcha*..." \
  --filter "@theseed/infra*..." \
  run build
uv sync






# Cloud Mfe


# Devprod Golink
./devprod/golink/proto/build.sh go py

# Newtype Guiproxy




















