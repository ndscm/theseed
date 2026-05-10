#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

# Convenience wrapper for running the bootstrap rule directly.
bazel run :bootstrap
