#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

# KNOWN ISSUE!!!
#
# git-rebase doesn't check untracked new file. If a new file is generated
# during the sanitization process, create a separate commit for the new file
# during the sanitizing rebase process. And carefully apply it (with rebase
# fixup) to the proper commit.

# Go

bazel run @rules_go//go -- mod tidy

# Python
if [[ "$(uname)" == "Darwin" ]]; then
  uv pip freeze --color never >./requirements_darwin.txt
else
  uv pip freeze --color never >./requirements.txt
fi
bazel run //seed/devprod/python/modules_mapping:generate
bazel run //:gazelle_python_manifest.update

# TypeScript
export ELECTRON_GET_USE_PROXY=1
# Install all node_modules for every package in the pnpm workspace.
bazel run @pnpm//:pnpm -- --dir $PWD install

# Gazelle
# Must run gazelle build file generator after all generators
bazel run //:gazelle

# Bazel
# Must tidy bazel mods after gazelle
bazel mod tidy
