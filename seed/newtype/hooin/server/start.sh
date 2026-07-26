#!/bin/bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

set +x
openid_client_secret="$(
  age -d \
    -i "$(ndscm secret --user get-path key.age)" \
    "$(ndscm secret --user get-path seed/newtype/hooin/server/OPENID_CLIENT_SECRET.age)"
)"
# Hand the decrypted secret to the server over a pipe fd. The pipe is
# single-use: once the server reads it, the plaintext is drained and no other
# process can re-read it from /proc/$$/fd. unset drops the shell's own copy so
# the cleartext doesn't linger in this process's memory for the server's life.
exec {openid_client_secret_fd}< <(printf "%s" "${openid_client_secret}")
unset openid_client_secret
set -x

bazel run //seed/newtype/hooin/server -- \
  --openid_client_id "hooin-dev-${ND_USER_HANDLE}" \
  --openid_client_secret_file "/proc/$$/fd/${openid_client_secret_fd}" \
  --verbose
