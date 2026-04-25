#!/usr/bin/env bash
set -eux
set -o pipefail

exec "/opt/webhook/webhook-relay" "$@"
