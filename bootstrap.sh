#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

# Monorepo
bazel run @pnpm//:pnpm -- --dir $PWD install
uv sync






# Newtype Guiproxy












