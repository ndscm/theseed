# sopsflag

Package `sopsflag` defines a secret flag whose value points at a SOPS-encrypted
file (the `sops:` file scheme). The flag decrypts the file in-process on `Load`,
via `github.com/getsops/sops`, and returns the plaintext secret.

## Not recommended

This package is no longer recommended for new code.

Decrypting SOPS in-process links `github.com/getsops/sops` into the binary,
which pulls in a large tree of unrelated dependencies — every KMS/key backend
SOPS supports (AWS KMS, GCP KMS, Azure Key Vault, HashiCorp Vault, age, PGP,
etc.) and their transitive SDKs. A binary that only needs to read one decrypted
secret ends up carrying all of that, bloating build times, binary size, and the
dependency attack surface.

## Preferred alternative

Decrypt the secret _outside_ the binary and hand it in as already-plaintext, so
the binary links none of the SOPS machinery. A convenient pattern is to decrypt
in the launch script and pass the plaintext over a single-use pipe fd via
`seedflag.DefineSecret`'s `_file` flag:

```bash
plain_secret="$(sops -d "secret.sops.yaml")"
exec {plain_secret_fd}< <(printf "%s" "${plain_secret}")
unset plain_secret
binary --secret_file "/proc/$$/fd/${plain_secret_fd}"
```

The SOPS toolchain stays in the `sops` CLI process and is gone before the binary
starts. The pipe is single-use: once the binary reads it, the plaintext is
drained and cannot be re-read from `/proc/$$/fd`.
