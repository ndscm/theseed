#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

branch=${1}
nd_user_handle=${2:-"${ND_USER_HANDLE}"}
workstation=$(hostname -s)

if [[ ! -d "${HOME}/theseed/devcron/${branch}" ]]; then
  git --git-dir "${HOME}/theseed/theseed.git" \
    worktree add -b "devcron/${branch}" "${HOME}/theseed/devcron/${branch}" "${branch}"
fi

cd "${HOME}/theseed/devcron/${branch}"

git clean -fdX
git clean -fd
git reset --hard ${branch}
rsync -av --delete --exclude=.git --exclude=node_modules --exclude=.venv \
  "${HOME}/theseed/${branch}/." \
  "${HOME}/theseed/devcron/${branch}/."

if [ ! -z "$(git status --porcelain)" ]; then
  git add --all
  git commit -m "wip from ${workstation} at $(date +%Y-%m-%dT%H:%M:%S%z)"
fi

git push origin --force "devcron/${branch}:${nd_user_handle}/${branch}-${workstation}"
