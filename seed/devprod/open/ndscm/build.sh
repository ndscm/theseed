#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

rm -rf /tmp/open
rm -rf /tmp/rebuild
bazel run //seed/devprod/open/ndscm
