#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

release_tag="${1:-""}"

if [[ -z "${release_tag}" ]]; then
  release_tag=$(date -u "+v%y.%m.%d" | sed 's/\.0*/./g; s/v0*/v/')
fi

git fetch --all --prune

if [[ -z "$(git tag --list "${release_tag}")" ]]; then
  git checkout main
  git reset --hard origin/main

  touch BUILD.bazel
  go mod tidy -e
  bazel mod tidy
  bazel run //seed/devprod/rbe/bes/proto:bootstrap
  go mod tidy
  bazel mod tidy
  bazel build //seed/devprod/ndscm/cli
  git add --all

  if [[ -n "$(git status --porcelain)" ]]; then
    git commit -m "seed: ndscm: release ${release_tag}"
  fi

  git tag "${release_tag}"
  git push origin "${release_tag}"
else
  git checkout main
  git reset --hard "${release_tag}"
fi

bazel build --stamp //seed/devprod/ndscm/cli

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m | tr '[:upper:]' '[:lower:]')"

tar czvf "./ndscm-${os}-${arch}.tar.gz" -C ./bazel-bin/seed/devprod/ndscm/cli/ndscm_/ ndscm
