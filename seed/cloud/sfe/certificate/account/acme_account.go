package account

import (
	"crypto"
	"encoding/json/v2"
	"fmt"
	"os"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/registration"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

// See: https://github.com/go-acme/lego/blob/57c14f8d2ab22bcd3d3bc7e36827f2b833da001f/cmd/account.go
type AcmeAccount struct {
	Email        string                `json:"email"`
	Registration *acme.ExtendedAccount `json:"registration"`
	key          crypto.Signer
}

func (a *AcmeAccount) GetEmail() string {
	return a.Email
}

func (a *AcmeAccount) GetPrivateKey() crypto.Signer {
	return a.key
}

func (a *AcmeAccount) GetRegistration() *acme.ExtendedAccount {
	return a.Registration
}

var _ registration.User = &AcmeAccount{}

func LoadAcmeAccount(acmeAccountJson []byte, acmeAccountKey []byte) (*AcmeAccount, error) {
	acmeAccount := &AcmeAccount{}

	err := json.Unmarshal(acmeAccountJson, &acmeAccount)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	privateKey, err := certcrypto.ParsePEMPrivateKey(acmeAccountKey)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	acmeAccount.key = privateKey

	return acmeAccount, nil
}

func LoadLocalAcmeAccount(accountsHome string, acmeId string, acmeEmail string) (*AcmeAccount, error) {
	acmeAccountJsonPath := fmt.Sprintf("%s/%s/%s/account.json", accountsHome, acmeId, acmeEmail)
	acmeAccountKeyPath := fmt.Sprintf("%s/%s/%s/%s.key", accountsHome, acmeId, acmeEmail, acmeEmail)

	acmeAccountJsonBytes, err := os.ReadFile(acmeAccountJsonPath)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	acmeAccountKeyBytes, err := os.ReadFile(acmeAccountKeyPath)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	return LoadAcmeAccount(acmeAccountJsonBytes, acmeAccountKeyBytes)
}
