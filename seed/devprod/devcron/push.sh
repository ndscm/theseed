#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

monorepo_home="${1}"
user_handle="${2}"
focus="${3}"
workstation=$(hostname -s)

if [[ ! -d "${monorepo_home}/${user_handle}/devcron/${focus}" ]]; then
  {
    cd "${monorepo_home}/main"
    git worktree add \
      -b "${user_handle}/devcron/${focus}" \
      "${monorepo_home}/${user_handle}/devcron/${focus}" \
      "${user_handle}/dev/${focus}"
  }
fi

cd "${monorepo_home}/${user_handle}/devcron/${focus}"

git clean -fdX
git clean -fd
git reset --hard "${user_handle}/dev/${focus}"
rsync -av --delete \
  --exclude=.git \
  --exclude=.venv \
  --exclude=node_modules \
  "${monorepo_home}/${user_handle}/dev/${focus}/." \
  "${monorepo_home}/${user_handle}/devcron/${focus}/."

if [[ ! -z "$(git status --porcelain)" ]]; then
  git add --all
  git commit -m "wip from ${workstation} at $(date +%Y-%m-%dT%H:%M:%S%z)"
fi

git push origin --force "${user_handle}/devcron/${focus}:${user_handle}/devcron/${focus}/${workstation}"
