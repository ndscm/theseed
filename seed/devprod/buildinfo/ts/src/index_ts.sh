#!/usr/bin/env bash
set -euo pipefail

# cat bazel-out/stable-status.txt bazel-out/volatile-status.txt | awk '{ print "export const " $1 " = \"" $2 "\"" }'

cat bazel-out/stable-status.txt bazel-out/volatile-status.txt | awk '{
    key=$1;
    sub(/^[^ ]*[ ]+/, "");
    printf "export const %s = \"%s\";\n", key, $0
}'
