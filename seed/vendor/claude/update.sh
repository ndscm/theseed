#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="${1:-"2.1.177"}"

if [[ "${tag}" == "latest" ]]; then
  tag="$(curl -fsSL https://downloads.claude.ai/claude-code-releases/latest)"
fi

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/claude/claude.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/claude/bin/claude
chmod +x ./seed/vendor/claude/bin/claude
