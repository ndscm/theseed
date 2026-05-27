#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

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

ndscm bootstrap
