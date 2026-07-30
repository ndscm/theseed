package secret

// Provider decrypts secrets of a particular kind, identified by the file
// extension of the secret path (e.g. ".age"). Each backend registers a provider
// via Register, and the active one is selected per secret through GetProvider.
type Provider interface {
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
