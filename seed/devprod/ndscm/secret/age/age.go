package age

import (
	"os"
	"path/filepath"

	"github.com/ndscm/theseed/seed/devprod/ndscm/secret"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type AgeProvider struct{}

func (a *AgeProvider) Decrypt(worktreePath string, secretPath string) error {
	ageKeyFile := os.Getenv("AGE_KEY_FILE")
	if ageKeyFile == "" {
		ageKeyFile = filepath.Join(worktreePath, "key.age")
	}
	secretAbsPath := filepath.Join(worktreePath, secretPath)
	err := seedshell.PureRun("age", "-d", "-i", ageKeyFile, secretAbsPath)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

var _ secret.Provider = (*AgeProvider)(nil)
