package secret

// SecretSessionProvider caches unlocked secrets (e.g. a decrypted age key) for
// the lifetime of a login session, so the user is prompted to unlock them once
// rather than on every command. Each backend registers a provider via
// RegisterSessionProvider, and the active one is selected through
// UseSecretSession per the secret_session flag.
type SecretSessionProvider interface {
	// Save stores value under key in the session. It overwrites any value
	// previously stored under the same key.
	Save(key string, value []byte) error

	// Load returns the value previously stored under key. It reports an error
	// if no value is stored under key, or if the session cannot be read.
	Load(key string) ([]byte, error)
}
