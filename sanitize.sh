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





# Bazel
bazel mod tidy
