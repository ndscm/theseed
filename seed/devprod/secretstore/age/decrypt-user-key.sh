#!/usr/bin/env bash

# This script is meant to be sourced (. decrypt-user-key.sh) so it can export
# AGE_KEY_FILE into the caller's shell. It deliberately does not enable
# `set -eux`/`pipefail`, which would leak into and disrupt the caller's shell.
#
# The work runs inside a function so its temporaries — notably the decrypted
# private key — stay `local` and never leak into the caller's shell on any exit
# path; only AGE_KEY_FILE is exported.

_decrypt_user_key() {
  local key_path
  local decrypted_key
  local key_fd

  key_path="$(ndscm secret --user get-path key.age)" || return 1
  decrypted_key="$(age -d "${key_path}")" || return 1

  exec {key_fd}<<<"${decrypted_key}"

  export AGE_KEY_FILE="/proc/$$/fd/${key_fd}"
  printf "export AGE_KEY_FILE=%s\n" "${AGE_KEY_FILE}" >&2
}

# Propagate the status as `return` when sourced or `exit` when executed, drop the
# helper, and leave no residue behind ($? is expanded before eval runs, so the
# unset does not clobber it).
_decrypt_user_key
eval "unset -f _decrypt_user_key; return $? 2>/dev/null || exit $?"
