package openidjwt

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagOpenidDiscovery = seedflag.DefineString("openid_discovery", "", `OpenID provider configurations file (in JSON format).`)

const refreshCooldown = 10 * time.Minute

func loadPublicKey(jwk *openid.OpenidJwk) (crypto.PublicKey, error) {
	switch jwk.Kty {
	case "RSA":
		if jwk.N == "" || jwk.E == "" {
			return nil, seederr.WrapErrorf("RSA key missing n or e")
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
		if err != nil {
			return nil, seederr.WrapErrorf("failed to decode RSA n: %v", err)
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
		if err != nil {
			return nil, seederr.WrapErrorf("failed to decode RSA e: %v", err)
		}
		return &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: int(new(big.Int).SetBytes(eBytes).Int64()),
		}, nil
	case "EC":
		if jwk.Crv == "" || jwk.X == "" || jwk.Y == "" {
			return nil, seederr.WrapErrorf("EC key missing crv, x, or y")
		}
		var curve elliptic.Curve
		switch jwk.Crv {
		case "P-256":
			curve = elliptic.P256()
		case "P-384":
			curve = elliptic.P384()
		case "P-521":
			curve = elliptic.P521()
		default:
			return nil, seederr.WrapErrorf("unsupported EC curve: %s", jwk.Crv)
		}
		xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
		if err != nil {
			return nil, seederr.WrapErrorf("failed to decode EC x: %v", err)
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
		if err != nil {
			return nil, seederr.WrapErrorf("failed to decode EC y: %v", err)
		}
		return &ecdsa.PublicKey{
			Curve: curve,
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}, nil
	default:
		return nil, seederr.WrapErrorf("unsupported key type: %s", jwk.Kty)
	}
}

type OpenidDiscoveryEntry struct {
	Discovery string `json:"discovery"`
}

type OpenidDiscovery struct {
	Issuers []OpenidDiscoveryEntry `json:"issuers"`
}

func fetchOpenidConfiguration(ctx context.Context, discoveryUrl string) (*openid.OpenidConfiguration, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryUrl, nil)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, seederr.WrapErrorf("unexpected status %d from %s", response.StatusCode, discoveryUrl)
	}
	configurationBytes, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	configuration, err := openid.DecodeOpenidConfiguration(configurationBytes)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if configuration.JwksUri == "" {
		return nil, seederr.WrapErrorf("jwks_uri not found in discovery document at %s", discoveryUrl)
	}
	return configuration, nil
}

type OpenidIssuerJwks struct {
	certificates map[string]*x509.Certificate
	publicKeys   map[string]crypto.PublicKey
}

func fetchOpenidIssuerJwks(ctx context.Context, discoveryUrl string) (string, *OpenidIssuerJwks, error) {
	configuration, err := fetchOpenidConfiguration(ctx, discoveryUrl)
	if err != nil {
		return "", nil, seederr.Wrap(err)
	}

	jwksRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, configuration.JwksUri, nil)
	if err != nil {
		return "", nil, seederr.Wrap(err)
	}
	jwksResponse, err := http.DefaultClient.Do(jwksRequest)
	if err != nil {
		return "", nil, seederr.Wrap(err)
	}
	defer jwksResponse.Body.Close()
	if jwksResponse.StatusCode != http.StatusOK {
		return "", nil, seederr.WrapErrorf("unexpected status %d from %s", jwksResponse.StatusCode, configuration.JwksUri)
	}
	jwksBytes, err := io.ReadAll(io.LimitReader(jwksResponse.Body, 1<<20))
	if err != nil {
		return "", nil, seederr.Wrap(err)
	}
	jwks, err := openid.DecodeOpenidJwks(jwksBytes)
	if err != nil {
		return "", nil, seederr.Wrap(err)
	}

	result := &OpenidIssuerJwks{
		certificates: map[string]*x509.Certificate{},
		publicKeys:   map[string]crypto.PublicKey{},
	}
	for _, jwk := range jwks.Keys {
		if jwk.Kid == "" {
			continue
		}
		if len(jwk.X5c) > 0 {
			certDer, err := base64.StdEncoding.DecodeString(jwk.X5c[0])
			if err != nil {
				seedlog.Errorf("Failed to decode x5c for kid %s: %v", jwk.Kid, err)
				continue
			}
			cert, err := x509.ParseCertificate(certDer)
			if err != nil {
				seedlog.Errorf("Failed to parse certificate for kid %s: %v", jwk.Kid, err)
				continue
			}
			result.certificates[jwk.Kid] = cert
			continue
		}
		pubKey, err := loadPublicKey(&jwk)
		if err != nil {
			seedlog.Errorf("Failed to extract public key for kid %s: %v", jwk.Kid, err)
			continue
		}
		result.publicKeys[jwk.Kid] = pubKey
	}
	return configuration.Issuer, result, nil
}

type OpenidJwksStore struct {
	discoveryMutex sync.RWMutex
	discovery      OpenidDiscovery

	issuersMutex sync.RWMutex
	issuers      map[string]*OpenidIssuerJwks

	lastRefreshMutex sync.Mutex
	lastRefresh      time.Time
}

func (s *OpenidJwksStore) getDiscovery() OpenidDiscovery {
	s.discoveryMutex.RLock()
	defer s.discoveryMutex.RUnlock()
	return s.discovery
}

func (s *OpenidJwksStore) clearIssuers() {
	s.issuersMutex.Lock()
	defer s.issuersMutex.Unlock()
	s.issuers = map[string]*OpenidIssuerJwks{}
}

func (s *OpenidJwksStore) setIssuer(issuer string, issuerJwks *OpenidIssuerJwks) {
	s.issuersMutex.Lock()
	defer s.issuersMutex.Unlock()
	s.issuers[issuer] = issuerJwks
}

func (s *OpenidJwksStore) refreshIssuers() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	discovery := s.getDiscovery()
	s.clearIssuers()
	for _, entry := range discovery.Issuers {
		seedlog.Infof("Fetching openid configuration: %s", entry.Discovery)
		issuer, jwks, err := fetchOpenidIssuerJwks(ctx, entry.Discovery)
		if err != nil {
			seedlog.Warnf("Failed to fetch keys from provider %s: %v", entry.Discovery, err)
			continue
		}
		s.setIssuer(issuer, jwks)
	}
	return nil
}

func CreateOpenidJwksStore() (*OpenidJwksStore, error) {
	discoveryConfigPath := flagOpenidDiscovery.Get()
	discovery := OpenidDiscovery{}
	if discoveryConfigPath != "" {
		discoveryConfigBytes, err := os.ReadFile(discoveryConfigPath)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		err = json.Unmarshal(discoveryConfigBytes, &discovery)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	}
	store := &OpenidJwksStore{
		discovery: discovery,
		issuers:   map[string]*OpenidIssuerJwks{},
	}
	err := store.refreshIssuers()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return store, nil
}

func (s *OpenidJwksStore) getPublicKey(issuer string, kid string) (crypto.PublicKey, error) {
	s.issuersMutex.RLock()
	defer s.issuersMutex.RUnlock()
	issuerJwks, ok := s.issuers[issuer]
	if !ok {
		return nil, nil
	}
	cert, ok := issuerJwks.certificates[kid]
	if ok {
		now := time.Now()
		if now.Before(cert.NotBefore) {
			return nil, seederr.WrapErrorf("certificate for kid %s is not yet valid (NotBefore: %v)", kid, cert.NotBefore)
		}
		if now.After(cert.NotAfter) {
			return nil, seederr.WrapErrorf("certificate for kid %s has expired (NotAfter: %v)", kid, cert.NotAfter)
		}
		return cert.PublicKey, nil
	}
	pubKey, ok := issuerJwks.publicKeys[kid]
	if ok {
		return pubKey, nil
	}
	return nil, nil
}

func (s *OpenidJwksStore) throttledRefresh() {
	s.lastRefreshMutex.Lock()
	defer s.lastRefreshMutex.Unlock()
	if time.Since(s.lastRefresh) < refreshCooldown {
		return
	}
	err := s.refreshIssuers()
	if err != nil {
		seedlog.Warnf("Failed to refresh issuers: %v", err)
		return
	}
	s.lastRefresh = time.Now()
}

func (s *OpenidJwksStore) GetByKid(issuer string, kid string) (crypto.PublicKey, error) {
	key, err := s.getPublicKey(issuer, kid)
	if err != nil {
		return nil, err
	}
	if key != nil {
		return key, nil
	}
	s.throttledRefresh()
	key, err = s.getPublicKey(issuer, kid)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, seederr.WrapErrorf("public key not found for kid: %s", kid)
	}
	return key, nil
}
