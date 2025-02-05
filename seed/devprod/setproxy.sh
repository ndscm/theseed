#!/bin/sh

# Setup:
#     ln -s ${HOME}/theseed/main/seed/devprod/setproxy.sh ~/setproxy.sh
#
# Usage:
#   Set proxy:
#     . ~/setproxy.sh http://proxy:port
#   Clean proxy:
#     . ~/setprox.sh

proxy=${1:-""}

if [[ "$proxy" == "" ]]; then
    unset HTTPS_PROXY
    unset HTTP_PROXY
    unset https_proxy
    unset http_proxy
else
    export HTTPS_PROXY=$proxy
    export HTTP_PROXY=$proxy
    export https_proxy=$proxy
    export http_proxy=$proxy
fi
