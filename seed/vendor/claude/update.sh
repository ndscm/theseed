#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag=$(curl -fsSL https://downloads.claude.ai/claude-code-releases/latest)

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/claude/claude.dotslash.json" \
  --tag="${tag}" \
  --no-format \
  >./seed/vendor/claude/bin/claude
chmod +x ./seed/vendor/claude/bin/claude
