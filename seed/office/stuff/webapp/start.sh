#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

stack=${1:-"local"}

pnpm install

if [[ "$stack" == "local" ]]; then
  export LOGIN_SERVICE_SERVER="http://127.0.0.1:7883"
  export STUFF_SERVICE_SERVER="http://127.0.0.1:7883"
fi

npx react-router dev --host
