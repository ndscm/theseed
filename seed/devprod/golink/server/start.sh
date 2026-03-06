#!/bin/bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

webapp_dir=""
if [[ -d "${PWD}/devprod/golink/webapp/dist/client" ]]; then
  webapp_dir="${PWD}/devprod/golink/webapp/dist/client"
fi

bazel run //devprod/golink/server -- \
  --webapp "$webapp_dir" \
  --verbose
