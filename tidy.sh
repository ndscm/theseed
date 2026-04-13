#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

# KNOWN ISSUE!!!
#
# git-rebase doesn't check untracked new file. If a new file is generated
# during the sanitization process, create a separate commit for the new file
# during the sanitizing rebase process. And carefully apply it (with rebase
# fixup) to the proper commit.

if [[ ! "${SEED_TIDY_FOCUS:-}" ]]; then
  export SEED_TIDY_FOCUS="all"
fi

get_changed_files() {
  {
    git ls-files --others --exclude-standard
    git diff --cached --name-only
    git diff --name-only HEAD~1 2>/dev/null || true
  } | sort -u
}

changed_files=$(get_changed_files)
changed_count=$(printf '%s\n' "$changed_files" | grep -c . || true)

if [[ "$changed_count" -gt 20 ]]; then
  export SEED_TIDY_FULL=1
fi

# Filter changed files by regex pattern, one match per line.
grep_changes() {
  printf '%s\n' "$changed_files" | grep -E "$1" || true
}

# Go

if [[ "${SEED_TIDY_FOCUS}" == "go" || "${SEED_TIDY_FOCUS}" == "all" ]]; then
  # Watch:
  #   go.mod
  #   go.sum
  #   MODULE.bazel
  #   *.go
  #   **/BUILD.bazel
  go_changes=$(grep_changes '\.(go|mod|sum|bazel)$')
  if [[ "${SEED_TIDY_FULL:-}" || -n "$go_changes" ]]; then
    bazel run @rules_go//go -- mod tidy
  fi
fi

# Python

if [[ "${SEED_TIDY_FOCUS}" == "py" || "${SEED_TIDY_FOCUS}" == "all" ]]; then
  # Watch:
  #   pyproject.toml
  #   requirements.txt
  #   requirements_darwin.txt
  #   gazelle_python.yaml
  #   gazelle_python_modules_mapping_darwin.json
  #   gazelle_python_modules_mapping_linux.json
  #   uv.lock
  #   MODULE.bazel
  #   **/BUILD.bazel
  py_changes=$(grep_changes '\.(toml|txt|yaml|json|lock|bazel)$')
  if [[ "${SEED_TIDY_FULL:-}" || -n "$py_changes" ]]; then
    if [[ "$(uname)" == "Darwin" ]]; then
      uv pip freeze --color never >./requirements_darwin.txt
    else
      uv pip freeze --color never >./requirements.txt
    fi
    bazel run //seed/devprod/python/modules_mapping:generate
    bazel run //:gazelle_python_manifest.update
  fi
fi

# TypeScript

if [[ "${SEED_TIDY_FOCUS}" == "ts" || "${SEED_TIDY_FOCUS}" == "all" ]]; then
  # Watch:
  #   package.json
  #   pnpm-workspace.yaml
  #   pnpm-lock.yaml
  #   **/package.json
  ts_changes=$(grep_changes '\.(ts|tsx|json|yaml)$')
  if [[ "${SEED_TIDY_FULL:-}" || -n "$ts_changes" ]]; then
    export ELECTRON_GET_USE_PROXY=1
    bazel run @pnpm//:pnpm -- --dir "${PWD}" install
  fi
fi

# Gazelle
# Must run gazelle build file generator after all generators

changed_files=$(get_changed_files)
gazelle_changes=$(grep_changes '\.(go|py|bazel)$')
if [[ "${SEED_TIDY_FULL:-}" || -n "$gazelle_changes" ]]; then
  bazel run //:gazelle
fi

# Bazel
# Must tidy bazel mods after gazelle

changed_files=$(get_changed_files)
bazel_changes=$(grep_changes '\.(bazel)$')
if [[ "${SEED_TIDY_FULL:-}" || -n "$bazel_changes" ]]; then
  bazel mod tidy
fi
