package signedjwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"sync"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagSfeOpenidClientPrivateKey = seedflag.DefineSecret(
	"openid_client_private_key",
	"Client PEM Key for OpenID Connect Private Key JWT",
)

func parseRsaPrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, seederr.WrapErrorf("failed to decode PEM block from client key")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, seederr.WrapErrorf("client key is not an RSA private key, got %T", parsed)
	}
	return rsaKey, nil
}

var store struct {
	mutex sync.RWMutex

	initialized bool

	privateKey *rsa.PrivateKey
}

func loadCachedPrivateKey() (*rsa.PrivateKey, bool) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	return store.privateKey, store.initialized
}

func loadPrivateKey() (*rsa.PrivateKey, error) {
	privateKey, initialized := loadCachedPrivateKey()
	if initialized {
		return privateKey, nil
	}
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if store.initialized {
		return store.privateKey, nil
	}
	store.initialized = true
	privateKeyBytes, err := flagSfeOpenidClientPrivateKey.Load()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	privateKey = nil
	if len(privateKeyBytes) > 0 {
		privateKey, err = parseRsaPrivateKey(privateKeyBytes)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	}
	store.privateKey = privateKey
	return store.privateKey, nil
}

func WrapRsaTokenSource(clientId string, audience string) (*openid.RsaTokenSource, error) {
	privateKey, err := loadPrivateKey()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if privateKey == nil {
		return nil, seederr.WrapErrorf("no private key")
	}
	rsaTokenSource := openid.NewRsaTokenSource(
		clientId, []string{audience}, privateKey,
	)
	return rsaTokenSource, nil
}
