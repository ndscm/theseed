#!/bin/bash
set -eux
set -o pipefail

if [[ ! "$#" -eq 2 ]]; then
  echo -e "\e[33mUsage: nd review <feature-name> <ref>|<hash>\e[0m"
  exit 1
fi

branch=${1}
target=${2}

if [[ ! -z "$(git status --porcelain)" ]]; then
  echo -e "\e[31mWorkspace is dirty\e[0m"
  exit 1
fi

if [[ "$(git branch --show-current)" != dev* ]]; then
  echo -e "\e[31mWorkspace is not on the dev branch\e[0m"
  exit 1
fi
current="$(git branch --show-current)"

if ! git rev-parse --abbrev-ref --symbolic-full-name "${current}@{upstream}"; then
  echo -e "\e[31mTracking upstream is missing\e[0m"
  exit 1
fi
tracking="$(git rev-parse --abbrev-ref --symbolic-full-name ${current}@{upstream})"

if [[ ! -z "$(git rev-list --merges ${tracking}..HEAD)" ]]; then
  echo -e "\e[31mCurrent dev branch is not a pure branch (contains merge commit)\e[0m"
  exit 1
fi

if [[ -z "$(git rev-list --ancestry-path=${target} ${tracking}..HEAD)" ]]; then
  echo -e "\e[31mTarget ${target} is not on the current dev branch\e[0m"
  exit 1
fi

if git rev-parse --verify --quiet "change/${branch}"; then
  echo -e "\e[31mBranch change/${branch} already exists\e[0m"
  exit 1
fi

git branch "change/${branch}" "${target}"
git branch --set-upstream-to="${tracking}" "change/${branch}"
git branch --set-upstream-to="change/${branch}" "${current}"
echo "Created change request as ${1}"
