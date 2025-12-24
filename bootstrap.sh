#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

# Monorepo
export ELECTRON_GET_USE_PROXY=1
bazel run @pnpm//:pnpm -- --dir $PWD install
## Build all dependency packages of apps
bazel run @pnpm//:pnpm -- --dir $PWD recursive \
  --filter "@theseed/*haraka..." \
  --filter "@theseed/*proto..." \
  --filter "@theseed/devprod-buildinfo*..." \
  --filter "@theseed/infra*..." \
  run build
uv sync






# Cloud Mfe


# Newtype Guiproxy

















