#!/bin/bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

webapp_dir=""
if [[ -d "${PWD}/office/stuff/webapp/dist/client" ]]; then
  webapp_dir="${PWD}/office/stuff/webapp/dist/client"
fi

bazel run //office/stuff/server -- \
  --webapp "$webapp_dir" \
  --verbose
