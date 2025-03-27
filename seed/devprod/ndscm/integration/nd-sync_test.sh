#!/bin/bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

bazel build //devprod/ndscm/cli:ndscm
cp -f ./bazel-bin/devprod/ndscm/cli/ndscm_/ndscm ./devprod/ndscm/integration/ndscm

image=$(docker build --quiet --file ./devprod/ndscm/integration/Dockerfile.ubuntu ./devprod/ndscm/.)
(($?)) && exit $?

docker run --interactive --rm --tty --volume "/tmp/nd_test_apt_cache:/var/cache/apt/archives" "${image}" bash -c "$(
  cat <<TEST_ND_SYNC_SUCCESS
set -eux -o pipefail

# # Setup

eval "\$(ndscm --shell-eval shell)"
nd dev
git config user.name Nagi
git config user.email nagi@ndscm.com

# # Simulate developing

echo A > main.c
git add main.c && git commit -m first
echo 2 >> main.c
git add main.c && git commit -m second
echo 3 >> main.c
git add main.c && git commit -m third
echo 4 >> main.c
git add main.c && git commit -m forth

nd cut feature HEAD^^ && result=0 || result=\$?

sleep 2 # Force wait for new commit time
git checkout -b ongoing/feature origin/main
git cherry-pick origin/main..change/feature
git push --force origin ongoing/feature:main
git checkout dev
sleep 2 # Force wait for new commit time

# # Run test

nd sync

# # Verify result

[[ "\${result}" -eq 0 ]] && true || false
[[ "\$(git rev-parse --abbrev-ref --symbolic-full-name dev@{upstream})" == "base/dev" ]] && true || false
[[ "\$(git rev-parse --abbrev-ref --symbolic-full-name base/dev@{upstream})" == "origin/main" ]] && true || false
[[ "\$(git branch --show-current)" == "dev" ]] && true || false
[[ "\$(git rev-parse dev)" == "\$(git rev-parse HEAD)" ]] && true || false
[[ "\$(git rev-parse base/dev)" == "\$(git rev-parse HEAD^^)" ]] && true || false
[[ "\$(git rev-parse origin/main)" == "\$(git rev-parse base/dev)" ]] && true || false

TEST_ND_SYNC_SUCCESS
)"
