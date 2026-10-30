#!/usr/bin/env bash
set -euo pipefail

if [[ $(git rev-parse --is-inside-work-tree) ]]; then
  dirty="$(git status --porcelain | wc -l | tr -d ' ')"
  branch="$(git branch --show-current)"
  tag="$(git describe --tags --always --abbrev=16 --dirty --broken)"

  brief="${tag}"
  if [[ "${branch}" != "" && "${branch}" != "main" ]]; then
    brief="${brief}~${branch}"
  fi
  if [[ "${dirty}" != "0" ]]; then
    brief="${brief}~${dirty}"
  fi

  commit="$(git rev-parse HEAD)"
fi

printf 'STABLE_BUILD_TAG %s\n' "${tag:-"unknown"}"
printf 'STABLE_BUILD_BRANCH %s\n' "${branch:-"unknown"}"
printf 'STABLE_BUILD_DIRTY %s\n' "${dirty:-"unknown"}"
printf 'STABLE_BUILD_BRIEF %s\n' "${brief:-"unknown"}"
printf 'STABLE_GIT_COMMIT %s\n' "${commit:-"unknown"}"
