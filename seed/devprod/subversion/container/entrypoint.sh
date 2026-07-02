#!/usr/bin/env bash
set -euo pipefail

# Ensure the repository parent and auth-configuration directories exist
# on the (persistent) volume. We only chown on creation so that manual
# ownership changes survive restarts; everything served afterwards must
# be created as www-data (e.g. `podman exec --user www-data ...`).
if [[ ! -d "${SVN_PARENT_PATH}" ]]; then
  mkdir -p "${SVN_PARENT_PATH}"
  chown www-data:www-data "${SVN_PARENT_PATH}"
fi
if [[ ! -d "${SVN_CONF_PATH}" ]]; then
  mkdir -p "${SVN_CONF_PATH}"
  chown www-data:www-data "${SVN_CONF_PATH}"
fi

# Seed a conservative path-based authorization file granting authenticated
# users read-only access to everything. Write access is opt-in: an
# administrator edits this file on the volume to grant `rw` to specific
# users or repositories.
if [[ ! -f "${SVN_AUTHZ_FILE}" ]]; then
  cat >"${SVN_AUTHZ_FILE}" <<'EOF'
[/]
* = r
EOF
  chown www-data:www-data "${SVN_AUTHZ_FILE}"
fi

# Seed an empty HTTP basic-auth password file. No users exist until an
# administrator adds them out-of-band; Require valid-user means nobody
# can authenticate in the meantime.
if [[ ! -f "${SVN_AUTH_USER_FILE}" ]]; then
  : >"${SVN_AUTH_USER_FILE}"
  chown www-data:www-data "${SVN_AUTH_USER_FILE}"
  echo "NOTE: ${SVN_AUTH_USER_FILE} is empty; no users can authenticate yet." >&2
  echo "      Add a user with: htpasswd -B ${SVN_AUTH_USER_FILE} <username>" >&2
fi

# When TLS is requested, generate a self-signed certificate on the
# (persistent) volume once and pass the SVN_HTTPS define so the vhost
# turns on SSLEngine. The certificate survives restarts; delete it on the
# volume to force regeneration (e.g. after changing SVN_SERVER_NAME).
apache_defines=(-D FOREGROUND)
if [[ "${SVN_HTTPS}" == "true" ]]; then
  if [[ ! -f "${SVN_SSL_CERT_FILE}" || ! -f "${SVN_SSL_KEY_FILE}" ]]; then
    openssl req -x509 -newkey rsa:2048 -nodes \
      -keyout "${SVN_SSL_KEY_FILE}" \
      -out "${SVN_SSL_CERT_FILE}" \
      -days 3650 \
      -subj "/CN=${SVN_SERVER_NAME}"
    chown www-data:www-data "${SVN_SSL_CERT_FILE}" "${SVN_SSL_KEY_FILE}"
    chmod 600 "${SVN_SSL_KEY_FILE}"
  fi
  apache_defines+=(-D SVN_HTTPS)
fi

# Hand off to Apache in the foreground. We source the APACHE_* runtime
# variables (which the config references) and ensure their directories
# exist, then exec apache2 directly rather than via the apache2ctl
# wrapper. That makes httpd PID 1 so it receives SIGTERM itself and shuts
# down promptly, instead of the wrapper swallowing the signal until a
# SIGKILL timeout.
# envvars references unset variables ($SUFFIX etc.), so relax nounset
# while sourcing it.
set +u
source /etc/apache2/envvars
set -u
mkdir -p "${APACHE_RUN_DIR}"
mkdir -p "${APACHE_LOCK_DIR}"
exec apache2 "${apache_defines[@]}"
