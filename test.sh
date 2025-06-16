#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

# Test all commits:
# $ git rebase --interactive --exec "./test.sh"

# Test commits since a specific commit:
# $ git rebase --interactive --exec "./test.sh" <commit-hash-or-tag>

bazel build //...

bazel test //...
