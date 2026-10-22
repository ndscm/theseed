#!/usr/bin/env bash
# Rebuild the current branch's history from its orphan root using a rebase-todo
# produced by export.py, restoring each commit's original author and committer
# metadata. Runs against the git repository in the current working directory.
#
# Usage: rebuild.sh <rebase-todo>
set -eux
set -o pipefail

todo="${1:-""}"

if [[ -z "${todo}" ]]; then
  echo "usage: $0 <rebase-todo>" >&2
  exit 1
fi

# Resolve to an absolute path: git runs the sequence editor from the repo root,
# not from our caller's cwd.
todo="$(realpath "${todo}")"

# git invokes the sequence editor as `$GIT_SEQUENCE_EDITOR <generated-todo>`, so
# copying our todo over that path replaces git's generated todo with ours.
GIT_SEQUENCE_EDITOR="cp -- '${todo}'" git rebase --interactive --root
