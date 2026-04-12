#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

# Test all commits:
# $ git rebase --interactive --exec "./sanitize.sh"

# Test commits since a specific commit:
# $ git rebase --interactive --exec "./sanitize.sh" <commit-hash-or-tag>

# KNOWN ISSUE!!!
#
# git-rebase doesn't check untracked new file. If a new file is generated
# during the sanitization process, create a separate commit for the new file
# during the sanitizing rebase process. And carefully apply it (with rebase
# fixup) to the proper commit.

# Returns the union of untracked files, staged changes, and last committed
# files. One path per line, deduplicated. Handles spaces and UTF-8 in paths
# by using newlines as separators internally and NULL-separating for xargs.
get_changed_files() {
  {
    git ls-files --others --exclude-standard
    git diff --cached --name-only
    git diff --name-only HEAD~1 -- . 2>/dev/null || true
  } | sort -u
}

changed_files=$(get_changed_files)
changed_count=$(printf '%s\n' "$changed_files" | grep -c . || true)

if [[ "$changed_count" -gt 10 ]]; then
  export SEED_SANITIZE_FULL=1
fi

# Filter changed files by regex pattern, one match per line.
grep_changes() {
  printf '%s\n' "$changed_files" | grep -E "$1" || true
}

# CC
if [[ "${SEED_SANITIZE_FULL:-}" ]]; then
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

# Go

bazel run @rules_go//go -- mod tidy

# Java
if [[ "${SEED_SANITIZE_FULL:-}" ]]; then
  find . -name "*.java" -print0 | xargs -0 clang-format -i
else
  java_changes=$(grep_changes '\.java$')
  if [[ -n "$java_changes" ]]; then
    printf '%s\n' "$java_changes" | tr '\n' '\0' | xargs -0 clang-format -i
  fi
fi

# Python
if [[ "${SEED_SANITIZE_FULL:-}" ]] || [[ -n $(grep_changes '(^|/)pyproject\.toml$') ]]; then
  if [[ "$(uname)" == "Darwin" ]]; then
    uv pip freeze --color never >./requirements_darwin.txt
  else
    uv pip freeze --color never >./requirements.txt
  fi
  bazel run //seed/devprod/python/modules_mapping:generate
  bazel run //:gazelle_python_manifest.update
fi

# TypeScript
if [[ "${SEED_FORMAT_FULL:-}" ]]; then
  bazel run //seed/devprod/format/prettier -- --write .
else
  prettier_changes=$(grep_changes '\.(ts|tsx|js|jsx|mts|cts|mjs|cjs|css|scss|less|json|yaml|yml|md|mdx|html|vue|svelte|graphql|gql)$')
  if [[ -n "$prettier_changes" ]]; then
    printf '%s\n' "$prettier_changes" | tr '\n' '\0' | xargs -0 bazel run //seed/devprod/format/prettier -- --write
  fi
fi

# Gazelle
# Must run gazelle build file generator after all generators
bazel run //:gazelle

# Bazel
# Must tidy bazel mods after gazelle
bazel mod tidy
