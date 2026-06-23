#!/usr/bin/env bash
set -eux
set -o pipefail

in_file="${1}"
out_file="${2}"

content=""
if [[ -f "${in_file}" ]]; then
  content="$(cat "${in_file}")"
fi

# JSON-escape the content: backslash first, then the quote and the control
# characters that appear in a Bazel workspace status file.
content="${content//\\/\\\\}"
content="${content//\"/\\\"}"
content="${content//$'\t'/\\t}"
content="${content//$'\r'/\\r}"
content="${content//$'\n'/\\n}"

# Emit the Bazel workspace status as a TypeScript module so tsc compiles the
# content straight into the output, leaving no runtime *.txt dependency.
printf 'export default "%s"\n' "${content}" >"${out_file}"
