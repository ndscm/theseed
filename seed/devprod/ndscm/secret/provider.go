package secret

// Provider decrypts secrets of a particular kind, identified by the file
// extension of the secret path (e.g. ".age"). Each backend registers a provider
// via Register, and the active one is selected per secret through GetProvider.
type Provider interface {
	// Keygen creates a new secret key and its matching recipients material in
	// worktreePath, the root of the secret worktree. It refuses to overwrite an
	// existing key. As with Decrypt, the private key must never land in a Go
	// string or buffer inside this process.
	Keygen(worktreePath string) error

	// Encrypt reads plaintext from stdin and writes the encrypted secret to
	// secretPath — relative to worktreePath — using the recipients material in
	// worktreePath, the root of the secret worktree.
	//
	// As with Decrypt, implementations should avoid loading the plaintext into
	// ndscm's own memory. Prefer streaming stdin straight into the underlying
	// tool (e.g. hand stdin to the age process), so the secret never lands in a
	// Go string or buffer inside this process.
	Encrypt(worktreePath string, secretPath string) error

	// Decrypt decrypts the secret at secretPath — relative to worktreePath —
	// and writes its plaintext to stdout. worktreePath is the root of the
	// secret worktree, so providers can locate both the secret and sibling
	// material (e.g. the age key) relative to it.
	//
	// Implementations should avoid loading the decrypted plaintext into
	// ndscm's own memory. Prefer streaming it straight to stdout from the
	// underlying tool (e.g. hand stdout to the age process), so the secret
	// never lands in a Go string or buffer inside this process.
	Decrypt(worktreePath string, secretPath string) error
}
