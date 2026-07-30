package secret

import (
	"path/filepath"
	"sync"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

var secretRegistryMutex sync.Mutex
var secretRegistry = map[string]Provider{}

// Register associates provider with a secret file extension (e.g. ".age").
func Register(extension string, provider Provider) {
	secretRegistryMutex.Lock()
	defer secretRegistryMutex.Unlock()
	secretRegistry[extension] = provider
}

// GetProvider returns the provider registered for secretPath's file extension.
func GetProvider(secretPath string) (Provider, error) {
	extension := filepath.Ext(secretPath)
	secretRegistryMutex.Lock()
	defer secretRegistryMutex.Unlock()
	provider, ok := secretRegistry[extension]
	if !ok {
		return nil, seederr.WrapErrorf("no secret provider registered for extension %q", extension)
	}
	return provider, nil
}
