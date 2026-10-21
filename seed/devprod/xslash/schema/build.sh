#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

pnpm dlx ts-json-schema-generator --path 'xslash-compat.ts' --type 'XslashCompat' --out 'xslash-compat.schema.json'
pnpm dlx ts-json-schema-generator --path 'xslash-entry.ts' --type 'XslashEntry' --out 'xslash-entry.schema.json'
pnpm dlx ts-json-schema-generator --path 'xslash-skeleton.ts' --type 'XslashSkeleton' --out 'xslash-skeleton.schema.json'
