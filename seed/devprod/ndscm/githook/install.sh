#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

# core.hooksPath redirects git to a custom hooks dir; git rev-parse --git-path
# hooks ignores it, so installing here would silently have no effect.
if custom_hooks_path=$(git config --get core.hooksPath); then
  printf "core.hooksPath is set to: %s\n" "${custom_hooks_path}" >&2
  printf "Unset it before installing: git config --unset core.hooksPath\n" >&2
  exit 1
fi

hooks_dir=$(git rev-parse --git-path hooks)

# A real git hooks dir ships a commit-msg.sample; without it, hooks_dir may
# have resolved to some unrelated system path we shouldn't write into.
if [[ ! -f "${hooks_dir}/commit-msg.sample" ]]; then
  printf "Can't find the git hooks directory: %s\n" "${hooks_dir}" >&2
  printf "No commit-msg.sample inside it.\n" >&2
  exit 1
fi

if [[ -f "${hooks_dir}/commit-msg" ]]; then
  printf "Can't install the commit-msg hook: one is already in place.\n" >&2
  printf "Remove it first: rm %s\n" "${hooks_dir}/commit-msg" >&2
  exit 1
fi

cp "./commit-msg" "${hooks_dir}/commit-msg"
chmod +x "${hooks_dir}/commit-msg"
