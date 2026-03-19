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

./bootstrap.sh

# CC
find . \
  -type d \( -name .git -o -name node_modules -o -name .venv -o -name dist \) -prune -o \
  -type f \( -name "*.c" -o -name "*.cc" -o -name "*.cpp" -o -name "*.h" \) -print0 |
  xargs -0 clang-format --verbose -i

# Go
# Must tidy go mod after all packages

bazel run @rules_go//go -- mod tidy

# Java
clang-format -i $(find . -name "*.java")

# Python
if [[ "$(uname)" == "Darwin" ]]; then
  uv pip freeze >./requirements_darwin.txt
else
  uv pip freeze >./requirements.txt
fi
bazel run //seed/devprod/python/modules_mapping:generate
bazel run //:gazelle_python_manifest.update

# TypeScript
bazel build //:node_modules/prettier
./bazel-bin/node_modules/prettier/bin/prettier.cjs --write .

# Gazelle
# Must run gazelle build file generator after all generators
bazel run //:gazelle

# Bazel
# Must tidy bazel mods after gazelle
bazel mod tidy
