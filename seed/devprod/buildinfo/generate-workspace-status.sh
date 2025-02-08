#!/usr/bin/env bash
set -euo pipefail

# Build Info

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

echo STABLE_BUILD_TAG ${tag}
echo STABLE_BUILD_BRANCH ${branch}
echo STABLE_BUILD_DIRTY ${dirty}
echo STABLE_BUILD_BRIEF ${brief}
echo STABLE_GIT_COMMIT $(git rev-parse HEAD)
