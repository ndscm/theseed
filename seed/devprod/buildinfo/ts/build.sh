#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

bazel build --output_groups=+types //seed/devprod/buildinfo/ts:ts_project
mkdir -p ./seed/devprod/buildinfo/ts/dist/
rsync -av ./bazel-bin/seed/devprod/buildinfo/ts/dist/. ./seed/devprod/buildinfo/ts/dist/.
