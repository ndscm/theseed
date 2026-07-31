package secret

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var secretRegistryMutex sync.Mutex
var secretRegistry = map[string]Provider{}
var secretExtensions = map[string]string{}

// Register associates provider with a named secret backend and the file
// extensions (e.g. ".age") that select it. Each extension must start with a dot.
func Register(name string, extensions []string, provider Provider) {
	secretRegistryMutex.Lock()
	defer secretRegistryMutex.Unlock()
	secretRegistry[name] = provider
	for _, extension := range extensions {
		if !strings.HasPrefix(extension, ".") {
			seedlog.Warnf("secret extension %q for provider %q does not start with a dot; prepending one", extension, name)
			extension = "." + extension
		}
		if existing, ok := secretExtensions[extension]; ok {
			seedlog.Warnf("secret extension %q already registered by provider %q; overriding with %q", extension, existing, name)
		}
		secretExtensions[extension] = name
	}
}

// GetProvider returns the provider registered under name.
func GetProvider(name string) (Provider, error) {
	secretRegistryMutex.Lock()
	defer secretRegistryMutex.Unlock()
	provider, ok := secretRegistry[name]
	if !ok {
		return nil, seederr.WrapErrorf("no secret provider registered with name %q", name)
	}
	return provider, nil
}

// InferProvider returns the provider registered for secretPath's file extension.
func InferProvider(secretPath string) (string, error) {
	extension := filepath.Ext(secretPath)
	secretRegistryMutex.Lock()
	defer secretRegistryMutex.Unlock()
	name, ok := secretExtensions[extension]
	if !ok {
		return "", seederr.WrapErrorf("no secret provider registered for extension %q", extension)
	}
	return name, nil
}
