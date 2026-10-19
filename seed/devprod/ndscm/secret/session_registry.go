package secret

import (
	"sync"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagSecretSession = seedflag.DefineStringList(
	"secret_session", []string{"keyctl"},
	"Secret session storage")

var sessionRegistryMutex sync.Mutex
var sessionRegistry = map[string](func() (SecretSessionProvider, error)){}

// RegisterSessionProvider associates provider with a named secret session backend.
func RegisterSessionProvider(name string, initialize func() (SecretSessionProvider, error)) {
	sessionRegistryMutex.Lock()
	defer sessionRegistryMutex.Unlock()
	sessionRegistry[name] = initialize
}

func getSessionProvider(name string) (func() (SecretSessionProvider, error), bool) {
	sessionRegistryMutex.Lock()
	defer sessionRegistryMutex.Unlock()
	initialize, ok := sessionRegistry[name]
	return initialize, ok
}

type singletonSecretSessionEntry struct {
	provider SecretSessionProvider
}

var singletonSecretSessionMutex sync.Mutex
var singletonSecretSession *singletonSecretSessionEntry

func UseSecretSession() (SecretSessionProvider, error) {
	singletonSecretSessionMutex.Lock()
	defer singletonSecretSessionMutex.Unlock()

	if singletonSecretSession != nil {
		return singletonSecretSession.provider, nil
	}

	for _, candidate := range flagSecretSession.Get() {
		if candidate == "" || candidate == "none" {
			singletonSecretSession = &singletonSecretSessionEntry{}
			break
		}

		initialize, ok := getSessionProvider(candidate)
		if !ok {
			continue
		}

		provider, err := initialize()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		if provider == nil {
			continue
		}
		singletonSecretSession = &singletonSecretSessionEntry{provider: provider}
		break
	}

	if singletonSecretSession == nil {
		singletonSecretSession = &singletonSecretSessionEntry{}
	}

	return singletonSecretSession.provider, nil
}
