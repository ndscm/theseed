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

# Go
# Must tidy go mod after all packages
bazel run @rules_go//go -- mod tidy

# Python
printf -- "--find-links https://download.pytorch.org/whl/torch\n" >./requirements.txt
uv pip freeze >>./requirements.txt
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
