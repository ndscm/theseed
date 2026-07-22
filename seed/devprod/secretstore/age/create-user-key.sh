#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

force="${1:-""}"

if [[ "${force}" == "t" || "${force}" == "--force" || "${force}" == "-f" ]]; then
  force="true"
fi

if [[ -f "$(ndscm secret --user get-path key.age)" && "${force}" != "true" ]]; then
  printf "User key already exists at %s\n" "$(ndscm secret --user get-path key.age)" >&2
  exit 1
fi

age-keygen -pq | age -p -a >"$(ndscm secret --user get-path key.age)"
age -d "$(ndscm secret --user get-path key.age)" |
  age-keygen -y \
    >"$(ndscm secret --user get-path recipients.txt)"
age-inspect --json "$(ndscm secret --user get-path key.age)"
