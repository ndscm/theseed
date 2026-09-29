#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

bazel build --stamp //seed/devprod/ndscm/cli
mkdir -p "${HOME}/.local/bin"
rm -f "${HOME}/.local/bin/ndscm"
cp -f ./bazel-bin/seed/devprod/ndscm/cli/ndscm_/ndscm "${HOME}/.local/bin/ndscm"
