#!/bin/bash
set -eux
set -o pipefail

exec "/opt/hooin/hooin-server" "$@"
