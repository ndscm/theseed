#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

secret_path="${1:-""}"
force="${2:-""}"

secret_path="${secret_path%/}"
secret_path="${secret_path%.age}"

if [[ "${force}" == "t" || "${force}" == "--force" || "${force}" == "-f" ]]; then
  force="true"
fi

if [[ "${secret_path}" == "" ]]; then
  printf "Usage: create-user-secret.sh <secret_path> [--force]\n" >&2
  exit 1
fi

set +x

secret_value="$(cat)"
secret_value="${secret_value#"${secret_value%%[![:space:]]*}"}"
secret_value="${secret_value%"${secret_value##*[![:space:]]}"}"

if [[ "${secret_value}" == "" ]]; then
  printf "error: no secret value provided on stdin\n" >&2
  exit 1
fi

secret_path="${secret_path}.age"

if [[ -f "$(ndscm secret --user get-path "${secret_path}")" && "${force}" != "true" ]]; then
  printf "User secret already exists at %s\n" "$(ndscm secret --user get-path "${secret_path}")" >&2
  exit 1
fi

mkdir -p "$(dirname "$(ndscm secret --user get-path "${secret_path}")")"
printf "%s" "${secret_value}" |
  age -e -R "$(ndscm secret --user get-path recipients.txt)" -a \
    >"$(ndscm secret --user get-path "${secret_path}")"

set -x

age-inspect --json "$(ndscm secret --user get-path "${secret_path}")"
