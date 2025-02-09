#!/usr/bin/env bash
set -euo pipefail

# Build Info

if [[ $(git rev-parse --is-inside-work-tree) ]]; then
  dirty=$(git status --porcelain | wc -l)
  branch=$(git branch --show-current)
  tag=$(git describe --tags)

  brief=${tag}
  if [ "$branch" != "main" ]; then
    brief="${brief} on ${branch}"
  fi
  if [ "$dirty" != "0" ]; then
    brief="${brief}~${dirty}"
  fi

  commit=$(git rev-parse HEAD)
fi

echo STABLE_BUILD_TAG ${tag:-"unknown"}
echo STABLE_BUILD_BRANCH ${branch:-"unknown"}
echo STABLE_BUILD_DIRTY ${dirty:-"unknown"}
echo STABLE_BUILD_BRIEF ${brief:-"unknown"}
echo STABLE_GIT_COMMIT ${commit:-"unknown"}








