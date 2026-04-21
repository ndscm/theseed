#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

# Returns the union of untracked files, staged changes, and last committed
# files. One path per line, deduplicated. Handles spaces and UTF-8 in paths
# by using newlines as separators internally and NULL-separating for xargs.
get_changed_files() {
  {
    git ls-files --others --exclude-standard
    git diff --cached --diff-filter=d --name-only
    git diff --diff-filter=d --name-only HEAD~1 2>/dev/null || true
  } | sort -u
}

changed_files=$(get_changed_files)
changed_count=$(printf '%s\n' "$changed_files" | grep -c . || true)

if [[ "$changed_count" -gt 10 ]]; then
  export SEED_FORMAT_FULL=1
fi

# Filter changed files by regex pattern, one match per line.
grep_changes() {
  printf '%s\n' "$changed_files" | grep -E "$1" || true
}

# CC
if [[ "${SEED_FORMAT_FULL:-}" ]]; then
  find . \
    -type d \( -name .git -o -name node_modules -o -name .venv -o -name dist \) -prune -o \
    -type f \( -name "*.c" -o -name "*.cc" -o -name "*.cpp" -o -name "*.h" \) -print0 |
    xargs -0 clang-format --verbose -i
else
  cc_changes=$(grep_changes '\.(c|cc|cpp|h)$')
  if [[ -n "$cc_changes" ]]; then
    printf '%s\n' "$cc_changes" | tr '\n' '\0' | xargs -0 clang-format --verbose -i
  fi
fi

# Java
if [[ "${SEED_FORMAT_FULL:-}" ]]; then
  find . -name "*.java" -print0 | xargs -0 clang-format -i
else
  java_changes=$(grep_changes '\.java$')
  if [[ -n "$java_changes" ]]; then
    printf '%s\n' "$java_changes" | tr '\n' '\0' | xargs -0 clang-format -i
  fi
fi

# TypeScript
if [[ "${SEED_FORMAT_FULL:-}" ]]; then
  bazel run //seed/devprod/format/prettier -- --write "$(pwd)"
else
  prettier_changes=$(grep_changes '\.(ts|tsx|js|jsx|mts|cts|mjs|cjs|css|scss|less|json|yaml|yml|md|mdx|html|vue|svelte|graphql|gql)$|(^|/)Jenkinsfile$')
  if [[ -n "$prettier_changes" ]]; then
    printf '%s\n' "$prettier_changes" | xargs -d '\n' realpath | tr '\n' '\0' | xargs -0 bazel run //seed/devprod/format/prettier -- --write
  fi
fi
