#!/bin/bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

docker run --interactive --rm --tty --volume "/tmp/nd_test_apt_cache:/var/cache/apt/archives" $(
  docker build --quiet --file ./devprod/ndscm/integration/Dockerfile.ubuntu ./devprod/ndscm/.
) bash -c "$(
  cat <<TEST_ND_CUT_SUCCESS
set -eux -o pipefail

# # Setup

source /home/ubuntu/ndscm/envsetup.sh
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

# # Run test

nd cut pending HEAD^^ && result=0 || result=\$?

# # Verify result

[[ "\${result}" -eq 0 ]] && true || false
[[ "\$(git rev-parse --abbrev-ref --symbolic-full-name dev@{upstream})" == "change/pending" ]] && true || false
[[ "\$(git rev-parse --abbrev-ref --symbolic-full-name change/pending@{upstream})" == "main" ]] && true || false
[[ "\$(git rev-parse change/pending)" == "\$(git rev-parse HEAD^^)" ]] && true || false

TEST_ND_CUT_SUCCESS
)"

docker run --interactive --rm --tty --volume "/tmp/nd_test_apt_cache:/var/cache/apt/archives" $(
  docker build --quiet --file ./devprod/ndscm/integration/Dockerfile.ubuntu ./devprod/ndscm/.
) bash -c "$(
  cat <<TEST_ND_CUT_ON_DIRTY_FAIL
set -eux -o pipefail

# # Setup

source /home/ubuntu/ndscm/envsetup.sh
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
echo 5 >> main.c

# # Run test

nd cut pending HEAD^^ && result=0 || result=\$?

# # Verify result

[[ ! "\${result}" -eq 0 ]] && true || false

TEST_ND_CUT_ON_DIRTY_FAIL
)"

docker run --interactive --rm --tty --volume "/tmp/nd_test_apt_cache:/var/cache/apt/archives" $(
  docker build --quiet --file ./devprod/ndscm/integration/Dockerfile.ubuntu ./devprod/ndscm/.
) bash -c "$(
  cat <<TEST_ND_CUT_ON_DEV_2_SUCCESS
set -eux -o pipefail

# # Setup

source /home/ubuntu/ndscm/envsetup.sh
nd dev
git config user.name Nagi
git config user.email nagi@ndscm.com

# # Simulate developing

nd dev 2
echo A > main.c
git add main.c && git commit -m first
echo 2 >> main.c
git add main.c && git commit -m second
echo 3 >> main.c
git add main.c && git commit -m third
echo 4 >> main.c
git add main.c && git commit -m forth

# # Run test

nd cut pending HEAD^^ && result=0 || result=\$?

# # Verify result

[[ "\${result}" -eq 0 ]] && true || false
[[ "\$(git rev-parse --abbrev-ref --symbolic-full-name dev-2@{upstream})" == "change/pending" ]] && true || false
[[ "\$(git rev-parse --abbrev-ref --symbolic-full-name change/pending@{upstream})" == "main" ]] && true || false
[[ "\$(git rev-parse change/pending)" == "\$(git rev-parse HEAD^^)" ]] && true || false

TEST_ND_CUT_ON_DEV_2_SUCCESS
)"

docker run --interactive --rm --tty --volume "/tmp/nd_test_apt_cache:/var/cache/apt/archives" $(
  docker build --quiet --file ./devprod/ndscm/integration/Dockerfile.ubuntu ./devprod/ndscm/.
) bash -c "$(
  cat <<TEST_ND_CUT_ON_MAIN_FAIL
set -eux -o pipefail

# # Setup

source /home/ubuntu/ndscm/envsetup.sh
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

# # Run test

cd /home/ubuntu/monorepo/main
nd cut pending HEAD^^ && result=0 || result=\$?

# # Verify result

[[ ! "\${result}" -eq 0 ]] && true || false

TEST_ND_CUT_ON_MAIN_FAIL
)"

docker run --interactive --rm --tty --volume "/tmp/nd_test_apt_cache:/var/cache/apt/archives" $(
  docker build --quiet --file ./devprod/ndscm/integration/Dockerfile.ubuntu ./devprod/ndscm/.
) bash -c "$(
  cat <<TEST_ND_CUT_NOT_ON_DEV_BRANCH_FAIL
set -eux -o pipefail

# # Setup

source /home/ubuntu/ndscm/envsetup.sh
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

# # Run test

git checkout -b test
nd cut pending HEAD^^ && result=0 || result=\$?

# # Verify result

[[ ! "\${result}" -eq 0 ]] && true || false

TEST_ND_CUT_NOT_ON_DEV_BRANCH_FAIL
)"

docker run --interactive --rm --tty --volume "/tmp/nd_test_apt_cache:/var/cache/apt/archives" $(
  docker build --quiet --file ./devprod/ndscm/integration/Dockerfile.ubuntu ./devprod/ndscm/.
) bash -c "$(
  cat <<TEST_ND_CUT_NO_TRACKING_UPSTREAM_FAIL
set -eux -o pipefail

# # Setup

source /home/ubuntu/ndscm/envsetup.sh
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

# # Run test

git branch --unset-upstream
nd cut pending HEAD^^ && result=0 || result=\$?

# # Verify result

[[ ! "\${result}" -eq 0 ]] && true || false

TEST_ND_CUT_NO_TRACKING_UPSTREAM_FAIL
)"

docker run --interactive --rm --tty --volume "/tmp/nd_test_apt_cache:/var/cache/apt/archives" $(
  docker build --quiet --file ./devprod/ndscm/integration/Dockerfile.ubuntu ./devprod/ndscm/.
) bash -c "$(
  cat <<TEST_ND_CUT_CONTAIN_MERGE_FAIL
set -eux -o pipefail

# # Setup

source /home/ubuntu/ndscm/envsetup.sh
nd dev
git config user.name Nagi
git config user.email nagi@ndscm.com

# # Simulate developing

echo A > main.c
git add main.c && git commit -m first
echo 2 >> main.c
git add main.c && git commit -m second
git checkout -b extra main
echo 3 > extra.c
git add extra.c && git commit -m third
echo 4 >> extra.c
git add extra.c && git commit -m forth
git checkout dev
git merge extra -m "merge"

# # Run test

nd cut pending HEAD && result=0 || result=\$?

# # Verify result

[[ ! "\${result}" -eq 0 ]] && true || false

TEST_ND_CUT_CONTAIN_MERGE_FAIL
)"

docker run --interactive --rm --tty --volume "/tmp/nd_test_apt_cache:/var/cache/apt/archives" $(
  docker build --quiet --file ./devprod/ndscm/integration/Dockerfile.ubuntu ./devprod/ndscm/.
) bash -c "$(
  cat <<TEST_ND_CUT_TARGET_NOT_ON_DEV_FAIL
set -eux -o pipefail

# # Setup

source /home/ubuntu/ndscm/envsetup.sh
nd dev
git config user.name Nagi
git config user.email nagi@ndscm.com

# # Simulate developing

echo A > main.c
git add main.c && git commit -m first
echo 2 >> main.c
git add main.c && git commit -m second
git checkout -b extra main
echo 3 > extra.c
git add extra.c && git commit -m third
echo 4 >> extra.c
git add extra.c && git commit -m forth
git checkout dev

# # Run test

nd cut pending extra && result=0 || result=\$?

# # Verify result

[[ ! "\${result}" -eq 0 ]] && true || false

TEST_ND_CUT_TARGET_NOT_ON_DEV_FAIL
)"

docker run --interactive --rm --tty --volume "/tmp/nd_test_apt_cache:/var/cache/apt/archives" $(
  docker build --quiet --file ./devprod/ndscm/integration/Dockerfile.ubuntu ./devprod/ndscm/.
) bash -c "$(
  cat <<TEST_ND_CUT_ALREADY_EXIST_FAIL
set -eux -o pipefail

# # Setup

source /home/ubuntu/ndscm/envsetup.sh
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

# # Run test

nd cut pending HEAD^^
nd cut pending HEAD^^ && result=0 || result=\$?

# # Verify result

[[ ! "\${result}" -eq 0 ]] && true || false

TEST_ND_CUT_ALREADY_EXIST_FAIL
)"
