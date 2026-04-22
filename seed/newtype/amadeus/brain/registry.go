package brain

import (
	"sync"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagBrain = seedflag.DefineString("brain", "", "Brain implementation (e.g. claudecli)")

// BrainEntry wraps a registered provider with exclusive initialization.
// Calling Initialize more than once returns an error; it is up to the
// caller (today: Conscious) not to re-initialize. The mutex also serializes
// concurrent Initialize attempts so the provider's Initialize runs
// exactly once on success, without holding brainRegistryMutex (which
// the provider may itself touch).
type BrainEntry struct {
	provider Brain

	initializedMutex sync.Mutex
	initialized      bool
}

func (e *BrainEntry) Initialize() error {
	e.initializedMutex.Lock()
	defer e.initializedMutex.Unlock()
	if e.initialized {
		return seederr.WrapErrorf("the brain is already initialized")
	}
	err := e.provider.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	e.initialized = true
	return nil
}

var brainRegistryMutex sync.Mutex
var brainRegistry = map[string]*BrainEntry{}

func Register(identifier string, provider Brain) {
	brainRegistryMutex.Lock()
	defer brainRegistryMutex.Unlock()
	brainRegistry[identifier] = &BrainEntry{provider: provider}
}

func getDefaultBrainEntry() (*BrainEntry, error) {
	brainRegistryMutex.Lock()
	defer brainRegistryMutex.Unlock()
	if len(brainRegistry) == 0 {
		return nil, seederr.WrapErrorf("no brain provider is registered")
	}
	brainIdentifier := flagBrain.Get()
	if brainIdentifier == "" {
		if len(brainRegistry) != 1 {
			return nil, seederr.WrapErrorf("multiple brain providers are registered, but no brain flag is set")
		}
		for k := range brainRegistry {
			brainIdentifier = k
		}
		seedlog.Infof("Brain: %s", brainIdentifier)
	}
	entry, ok := brainRegistry[brainIdentifier]
	if !ok {
		return nil, seederr.WrapErrorf("brain provider not found: %s", brainIdentifier)
	}
	return entry, nil
}

func DefaultBrain() (Brain, error) {
	entry, err := getDefaultBrainEntry()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	err = entry.Initialize()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return entry.provider, nil
}
