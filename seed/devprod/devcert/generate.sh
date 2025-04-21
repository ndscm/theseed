#!/bin/bash
set -eux
set -o pipefail
cd $(dirname "${BASH_SOURCE[0]}")

service=${1-"accounts"}
domain=${2-"ndscm.com"}

mkdir -p "${ND_USER_SECRET_HOME}/devcert"

if [[ ! -f "${ND_USER_SECRET_HOME}/devcert/ca.dev.${ND_USER_HANDLE}.key" ]]; then
  openssl genrsa \
    -out "${ND_USER_SECRET_HOME}/devcert/ca.dev.${ND_USER_HANDLE}.key" \
    4096
  openssl req -new -sha256 \
    -key "${ND_USER_SECRET_HOME}/devcert/ca.dev.${ND_USER_HANDLE}.key" \
    -subj "/C=US/ST=California/L=San Francisco/O=Ndscm Inc/OU=Devprod/CN=${ND_USER_HANDLE}@ndscm.com Dev CA" \
    -out "${ND_USER_SECRET_HOME}/devcert/ca.dev.${ND_USER_HANDLE}.csr"
  cat >"${ND_USER_SECRET_HOME}/devcert/ca.dev.${ND_USER_HANDLE}.ext" <<EOF
authorityKeyIdentifier = keyid:always, issuer:always
basicConstraints = critical, CA:TRUE
keyUsage = critical, cRLSign, keyCertSign
subjectKeyIdentifier = hash
EOF
  openssl x509 -req -days 3650 \
    -in "${ND_USER_SECRET_HOME}/devcert/ca.dev.${ND_USER_HANDLE}.csr" \
    -extfile "${ND_USER_SECRET_HOME}/devcert/ca.dev.${ND_USER_HANDLE}.ext" \
    -signkey "${ND_USER_SECRET_HOME}/devcert/ca.dev.${ND_USER_HANDLE}.key" \
    -out "${ND_USER_SECRET_HOME}/devcert/ca.dev.${ND_USER_HANDLE}.crt"
fi

openssl genrsa \
  -out "${ND_USER_SECRET_HOME}/devcert/${service}.dev.${ND_USER_HANDLE}.local.${domain}.key" \
  4096
openssl req -new -sha256 \
  -key "${ND_USER_SECRET_HOME}/devcert/${service}.dev.${ND_USER_HANDLE}.local.${domain}.key" \
  -subj "/C=US/ST=California/L=San Francisco/O=Ndscm Inc/OU=Devprod/CN=${service}.dev.${ND_USER_HANDLE}.local.${domain}" \
  -out "${ND_USER_SECRET_HOME}/devcert/${service}.dev.${ND_USER_HANDLE}.local.${domain}.csr"
cat >"${ND_USER_SECRET_HOME}/devcert/${service}.dev.${ND_USER_HANDLE}.local.${domain}.ext" <<EOF
authorityKeyIdentifier = keyid, issuer:always
basicConstraints = critical, CA:FALSE
extendedKeyUsage = serverAuth, clientAuth
keyUsage = critical, digitalSignature, keyEncipherment
subjectAltName=DNS:${service}.dev.${ND_USER_HANDLE}.local.${domain}
subjectKeyIdentifier = hash
EOF
openssl x509 -req -days 30 \
  -CAkey "${ND_USER_SECRET_HOME}/devcert/ca.dev.${ND_USER_HANDLE}.key" \
  -CA "${ND_USER_SECRET_HOME}/devcert/ca.dev.${ND_USER_HANDLE}.crt" \
  -in "${ND_USER_SECRET_HOME}/devcert/${service}.dev.${ND_USER_HANDLE}.local.${domain}.csr" \
  -extfile "${ND_USER_SECRET_HOME}/devcert/${service}.dev.${ND_USER_HANDLE}.local.${domain}.ext" \
  -out "${ND_USER_SECRET_HOME}/devcert/${service}.dev.${ND_USER_HANDLE}.local.${domain}.crt"
