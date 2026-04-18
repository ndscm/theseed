#!/bin/bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../../..

bazel run //seed/newtype/amadeus/server -- \
  --verbose
