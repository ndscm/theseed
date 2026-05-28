#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

handoff=${1:-"${HOME}/Downloads/KurisuDesign-handoff.zip"}

if [[ ! -f "${handoff}" ]]; then
  echo "Handoff file not found: ${handoff}"
  exit 1
fi

rm -rf ./.design/
mkdir -p ./.design/
unzip "${handoff}" -d ./.design/
rm "${handoff}"
