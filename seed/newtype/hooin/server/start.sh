#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

bazel run //seed/newtype/hooin/server -- \
  --static_team_file=$(pwd)/seed/newtype/hooin/container/team.json \
  --verbose
