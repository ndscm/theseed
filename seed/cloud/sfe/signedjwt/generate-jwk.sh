#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

# Generate an RSA-4096 key pair for the BFF (Keycloak client).
#
# - private_key.pem: the client keeps this and uses it to sign the JWT that
#   authenticates it to Keycloak (private_key_jwt client authentication).
# - public_key.pem:  imported into the Keycloak client configuration so
#   Keycloak can verify the JWT signature.

private_key="private_key.pem"
public_key="public_key.pem"

# RSA-4096 private key in PKCS#8 PEM format.
openssl genpkey \
  -algorithm RSA \
  -pkeyopt rsa_keygen_bits:4096 \
  -out "${private_key}"
chmod 600 "${private_key}"

# Public key in PEM format (SubjectPublicKeyInfo).
openssl pkey \
  -in "${private_key}" \
  -pubout \
  -out "${public_key}"

echo "wrote ${private_key} (keep secret) and ${public_key} (import into Keycloak)"
