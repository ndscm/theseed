#!/bin/bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

sudo cp -f "${ND_USER_SECRET_HOME}/devcert/ca.dev.${ND_USER_HANDLE}.crt" /usr/local/share/ca-certificates/
sudo update-ca-certificates
