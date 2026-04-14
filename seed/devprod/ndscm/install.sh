#!/bin/bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../..

bazel build //seed/devprod/ndscm/cli
mkdir -p ${HOME}/.local/bin
cp -f bazel-bin/seed/devprod/ndscm/cli/ndscm_/ndscm ${HOME}/.local/bin/ndscm
