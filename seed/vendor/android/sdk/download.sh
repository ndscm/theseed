#!/usr/bin/env bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

curl -O -L https://dl.google.com/android/repository/repository2-3.xml
