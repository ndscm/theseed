#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

find ./seed/newtype/kurisu/webapp/locales/en/ -type f -name "*.json" -delete
(cd ./seed/newtype/kurisu/webapp/ && pnpm install && npx i18next-cli extract)

claude \
  --allowedTools "Edit(seed/newtype/kurisu/webapp/locales/es/**),Write(seed/newtype/kurisu/webapp/locales/es/**)" \
  -p "Please reference ./seed/newtype/kurisu/webapp/locales/en/ and translate the missing items in ./seed/newtype/kurisu/webapp/locales/es/ to Spanish"
