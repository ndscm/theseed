# ageflag

Package `ageflag` defines a secret flag whose value points at an `age`-encrypted
file (the `age:` file scheme). The flag decrypts the file in-process on `Load`,
using the identities read from `--age_key_file`, and returns the plaintext
secret.

## Not recommended

This package is no longer recommended for new code.

To decrypt in-process, the binary must hold the `age` private key (identity) in
its own memory for as long as it can decrypt secrets. That widens the attack
surface: the private key — which unlocks _every_ secret encrypted to it, not
just the one being loaded — is resident in the process and readable via
`/proc/<pid>/mem`, core dumps, or a compromise of the running binary.

## Preferred alternative

Decrypt the secret _outside_ the binary and hand it in as already-plaintext, so
the binary never sees the private key at all. A convenient pattern is to decrypt
in the launch script and pass the plaintext over a single-use pipe fd via
`seedflag.DefineSecret`'s `_file` flag:

```bash
plain_secret="$(age -d -i "key.age" "secret.age")"
exec {plain_secret_fd}< <(printf "%s" "${plain_secret}")
unset plain_secret
binary --secret_file "/proc/$$/fd/${plain_secret_fd}"
```

The private key stays with the `age` CLI process and is gone before the binary
starts. The pipe is single-use: once the binary reads it, the plaintext is
drained and cannot be re-read from `/proc/$$/fd`.
