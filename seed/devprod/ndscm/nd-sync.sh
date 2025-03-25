#!/bin/bash
set -eux
set -o pipefail

if [[ ! "$#" -eq 0 ]]; then
  echo -e "\e[33mUsage: nd sync\e[0m"
  exit 1
fi

if [[ -z "${ND_MONOREPO_HOME+x}" ]]; then
  echo -e "\e[31mND_MONOREPO_HOME is not set\e[0m"
  return 1
fi

if [[ ! -z "$(git status --porcelain)" ]]; then
  echo -e "\e[31mWorkspace is dirty\e[0m"
  exit 1
fi

target="$(git branch --show-current)"
worktree="$(realpath -s --relative-to "${ND_MONOREPO_HOME}" $(git rev-parse --show-toplevel))"

if [[ "${target}" != "${worktree}" ]]; then
  echo -e "\e[31mWorktree (${worktree}) and current branch (${target}) mismatch\e[0m"
  exit 1
fi

# # Iterate changes tree

upstreams=("${target}")
iter="${target}"
while [[ "${iter}" != "base/${target}" ]]; do
  inspect=$(git rev-parse --abbrev-ref --symbolic-full-name "${iter}@{upstream}")
  if [[ -z "${inspect+x}" ]]; then
    echo -e "\e[31mTracking upstream is missing for ${iter}\e[0m"
    exit 1
  fi
  if [[ "${inspect}" == origin/* ]]; then
    echo -e "\e[31mTracking upstream to remote at ${inspect} branch.\e[0m"
    exit 1
  fi
  upstreams+=("${inspect}")
  iter="${inspect}"
done

# # Fetch upstream changes

git fetch --all --prune

# # Rebase dev branch
echo "upstreams: ${upstreams[@]}"

incoming=($(git rev-list "base/${target}..origin/main"))
# TODO(nagi): resolve merge commits
for i in $(seq $(("${#incoming[@]}" - 1)) -1 0); do
  echo -e "\e[34mRebasing to ${incoming[i]}.\e[0m"
  git checkout base/${target}
  git rebase "${incoming[i]}"
  for b in $(seq $(("${#upstreams[@]}" - 2)) -1 0); do
    git checkout "${upstreams[b]}"
    git pull --rebase
  done
  echo -e "\e[32mRebased to ${incoming[i]}.\e[0m"
done

# # Cleanup local change branches

iter="${target}"
while [[ "${iter}" != "base/${target}" ]]; do
  inspect=$(git rev-parse --abbrev-ref --symbolic-full-name "${iter}@{upstream}")
  if [[ -z "${inspect+x}" ]]; then
    echo -e "\e[31mTracking upstream is missing for ${iter}\e[0m"
    exit 1
  fi
  if [[ "${inspect}" == origin/* ]]; then
    echo -e "\e[31mTracking upstream to remote at ${inspect} branch.\e[0m"
    exit 1
  fi
  if [[ "${inspect}" == "base/${target}" ]]; then
    break
  fi
  next=$(git rev-parse --abbrev-ref --symbolic-full-name "${inspect}@{upstream}")
  if [[ "$(git rev-parse ${inspect})" == "$(git rev-parse ${next})" ]]; then
    if [[ "${inspect}" == change/* ]]; then
      git branch -d "${inspect}"
      git branch --set-upstream-to "${next}" "${iter}"
      echo -e "\e[33mRemoved ${inspect} \e[0m"
    else
      echo -e "\e[33mWARNING: Branch ${inspect} should not be here. Please remove with
    git branch -d ${inspect} && git branch -u ${next} ${iter}\e[0m"
      iter="${inspect}"
    fi
  else
    iter="${inspect}"
  fi
done
