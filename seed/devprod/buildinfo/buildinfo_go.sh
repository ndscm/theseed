#!/usr/bin/env bash
set -euo pipefail

echo "package buildinfo"
cat bazel-out/stable-status.txt bazel-out/volatile-status.txt | awk '{print "var " $1 " string = \"" $2 "\""}'
