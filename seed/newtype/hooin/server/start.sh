#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

set +x

age_key_file="${AGE_KEY_FILE:-""}"
if [[ -z "${age_key_file}" ]]; then
  age_key_file="$(ndscm secret --user get-path key.age)"
fi

openid_client_secret="$(
  age -d -i "${age_key_file}" \
    "$(ndscm secret --user get-path seed/newtype/hooin/server/OPENID_CLIENT_SECRET.age)"
)"
exec {openid_client_secret_fd}< <(printf "%s" "${openid_client_secret}")
unset openid_client_secret

steins_database_secret="$(
  age -d -i "${age_key_file}" \
    "$(ndscm secret --user get-path seed/newtype/steins/database/STEINS_DATABASE_SECRET.local.age)"
)"
exec {steins_database_secret_fd}< <(printf "%s" "${steins_database_secret}")
unset steins_database_secret

set -x

bazel run //seed/newtype/hooin/server -- \
  --openid_client_id "hooin-dev-${ND_USER_HANDLE}" \
  --openid_client_secret_file "/proc/$$/fd/${openid_client_secret_fd}" \
  --steins_database "postgres://127.0.0.1:5432/steins_local" \
  --steins_database_debug \
  --steins_database_login "steins_local" \
  --steins_database_secret_file "/proc/$$/fd/${steins_database_secret_fd}" \
  --verbose
