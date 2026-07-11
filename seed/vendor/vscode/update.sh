#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

curl -fsSL https://update.code.visualstudio.com/api/update/web-standalone/stable/latest | jq
