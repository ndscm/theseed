#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

pnpm install
npx @vscode/vsce package --no-dependencies
code --install-extension vscode-custom-formatter-0.0.0.vsix
