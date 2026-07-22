package ageflag

import (
	"bytes"
	"sync"

	"filippo.io/age"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagAgeKeyFile = seedflag.DefineFile(
	"age_key_file", "",
	"Path to the age key file",
)

var cachedIdentities []age.Identity
var cachedIdentitiesErr error
var onceLoadIdentities sync.Once

func LoadIdentities() ([]age.Identity, error) {
	onceLoadIdentities.Do(func() {
		keyBytes, err := flagAgeKeyFile.Load()
		if err != nil {
			cachedIdentitiesErr = seederr.Wrap(err)
			return
		}
		identities, err := age.ParseIdentities(bytes.NewReader(keyBytes))
		if err != nil {
			cachedIdentitiesErr = seederr.Wrap(err)
			return
		}
		cachedIdentities = identities
	})
	return cachedIdentities, cachedIdentitiesErr
}
