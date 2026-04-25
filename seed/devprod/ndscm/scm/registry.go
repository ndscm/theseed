package scm

import (
	"sync"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

// The default SCM provider is "ndscm", which is used to force the user to set git explicitly.
// User can also provide empty value to use the default provider if there is only one provider registered.
var flagScm = seedflag.DefineString("scm", "ndscm", "The SCM provider to use (e.g. git)")

func ScmName() (string, error) {
	// TODO(nagi): remove this flag getter after all the checks are migrated.
	scm := flagScm.Get()
	switch scm {
	case "git":
		return "git", nil
	case "":
		return "", nil
	}
	return "", seederr.WrapErrorf("scm is unsupported: %v", scm)
}

type ProviderEntry struct {
	provider Provider

	initializedMutex sync.Mutex
	initialized      bool
}

func (e *ProviderEntry) Initialize() error {
	e.initializedMutex.Lock()
	defer e.initializedMutex.Unlock()
	if e.initialized {
		return seederr.WrapErrorf("the scm is already initialized")
	}
	err := e.provider.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	e.initialized = true
	return nil
}

var scmRegistryMutex sync.Mutex
var scmRegistry = map[string]*ProviderEntry{}

func Register(identifier string, provider Provider) {
	scmRegistryMutex.Lock()
	defer scmRegistryMutex.Unlock()
	scmRegistry[identifier] = &ProviderEntry{provider: provider}
}

func getDefaultProviderEntry() (*ProviderEntry, error) {
	scmRegistryMutex.Lock()
	defer scmRegistryMutex.Unlock()
	if len(scmRegistry) == 0 {
		return nil, seederr.WrapErrorf("no scm provider is registered")
	}
	scmIdentifier := flagScm.Get()
	if scmIdentifier == "" {
		if len(scmRegistry) != 1 {
			return nil, seederr.WrapErrorf("multiple scm providers are registered, but no scm flag is set")
		}
		for k := range scmRegistry {
			scmIdentifier = k
		}
		seedlog.Infof("SCM: %s", scmIdentifier)
	}
	entry, ok := scmRegistry[scmIdentifier]
	if !ok {
		return nil, seederr.WrapErrorf("SCM provider not found: %s", scmIdentifier)
	}
	return entry, nil
}

func InitializeDefaultProvider() (Provider, error) {
	entry, err := getDefaultProviderEntry()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	err = entry.Initialize()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return entry.provider, nil
}

func GetIdentifier(scmProvider Provider) (string, error) {
	scmRegistryMutex.Lock()
	defer scmRegistryMutex.Unlock()
	for identifier, entry := range scmRegistry {
		if entry.provider == scmProvider {
			return identifier, nil
		}
	}
	return "", seederr.WrapErrorf("SCM provider not found")
}
