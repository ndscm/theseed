#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
    if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
        cat <<EOF >>"${HOME}/.managed_shrc.tmp"

## Shortcuts

function main { cd "\${ND_MONOREPO_HOME}/main"; }
function dev { cd "\${ND_MONOREPO_HOME}/dev"; }
EOF
    fi

    if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
        ln -s -f -n "${ND_MONOREPO_HOME}/main/seed/devprod/setproxy.sh" "${HOME}/.local/bin/setproxy"
    fi
fi
