#!/bin/bash
set -eux
set -o pipefail

git fetch upstream --prune
git fetch origin --prune

upstream_branches=$(git ls-remote --heads upstream | awk '{ sub("refs/heads/", ""); print $2 }')
for branch in ${upstream_branches}; do
  printf "Checking branch ${branch}\n"
  git push origin --force "refs/remotes/upstream/${branch}:refs/heads/${branch}"
done

origin_branches=$(git ls-remote --heads origin | awk '{ sub("refs/heads/", ""); print $2 }')
deprecated_branches=()
for branch in ${origin_branches}; do
  if git rev-parse --verify "refs/remotes/upstream/$branch" &>/dev/null; then
    printf "Skipping upstream branch ${branch}\n"
  elif [[ "${branch}" == theseed/* ]]; then
    printf "Skipping theseed branch ${branch}\n"
  else
    deprecated_branches+=("${branch}")
  fi
done

set +x
printf "Remove deprecated branches with:\n"
for branch in "${deprecated_branches[@]}"; do
  printf "git push origin --delete \"refs/heads/${branch}\"\n"
done
