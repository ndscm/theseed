#!/usr/bin/env bash
set -eux
set -o pipefail

port=${1:-"9000"}

powershell.exe "sudo netsh interface portproxy add v4tov4 listenport=${port} listenaddress=0.0.0.0 connectport=${port} connectaddress=$(hostname -I)"
