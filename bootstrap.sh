#!/usr/bin/env bash
# bootstrap.sh — Regenerate all derived/generated files in the monorepo worktree.
#
# These files are typically listed in .gitignore and are not committed to the
# repository. They must be regenerated after cloning or switching branches to
# restore a working development environment. This includes:
#   - Generated protobuf/gRPC stubs (Go, Python, TypeScript)
#   - Generated ORM source files (ent)
#   - node_modules and pre-built workspace packages
#   - Python virtual environment (.venv) via uv
#   - Any per-package generated database and proto artifacts
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

# Signal to build rules that we are running a full monorepo bootstrap,
# allowing rules to opt into bootstrap-specific behaviour.
export SEED_MONOREPO_BOOTSTRAP=1

# --- Generated source files (Bazel) ---
# Build all code-generated targets across the monorepo:
#   ent_full_srcs   — ent ORM generated Go sources
#   _go_connect_go  — Connect-Go RPC generated sources
#   _go_proto       — protobuf generated Go sources
#   _py_grpc        — protobuf generated Python gRPC stubs
#   _py_pb2         — protobuf generated Python message classes
#   _ts_proto       — protobuf generated TypeScript sources
# The extra output groups ensure that Go-generated sources and TypeScript
# declaration files (.d.ts) are written into the source tree.
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

# --- Node.js dependencies (pnpm) ---
export ELECTRON_GET_USE_PROXY=1
# Install all node_modules for every package in the pnpm workspace.
bazel run @pnpm//:pnpm -- --dir $PWD install
# Pre-build all library/shared packages that webapp targets depend on,
# so that local package imports resolve correctly during development.
bazel run @pnpm//:pnpm -- --dir $PWD recursive \
  --filter "@theseed/*-webapp^..." \
  run build

# --- Python virtual environment (.venv) ---
# Sync the project's Python dependencies declared in pyproject.toml into
# the local .venv managed by uv.
uv sync

# --- Per-package generated artifacts ---
# Some packages carry their own generation scripts that are not (yet) driven
# by Bazel rules.  Run them all uniformly:
#   database/build.sh — generates database migration helpers / query code
#   proto/build.sh    — generates additional proto bindings (Go + Python)
find . -type f -path "*/database/build.sh" -exec bash {} \;
find . -type f -path "*/proto/build.sh" -exec bash {} go py \;
