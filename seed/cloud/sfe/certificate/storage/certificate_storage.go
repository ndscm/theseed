package storage

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"golang.org/x/net/idna"
)

func checkCertificate(acmeCertificate *certificate.Resource) error {
	if acmeCertificate == nil {
		return seederr.WrapErrorf("acme certificate is nil")
	}
	if acmeCertificate.IssuerCertificate == nil {
		return seederr.WrapErrorf("acme certificate has no issuer certificate")
	}
	if acmeCertificate.Certificate == nil {
		return seederr.WrapErrorf("acme certificate has no public certificate")
	}
	if acmeCertificate.PrivateKey == nil {
		return seederr.WrapErrorf("acme certificate has no private key")
	}
	x509Certificates, err := certcrypto.ParsePEMBundle(acmeCertificate.Certificate)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(x509Certificates) == 0 {
		return seederr.WrapErrorf("acme certificate has no certificate in the bundle")
	}
	x509Certificate := x509Certificates[0]
	alive := time.Now().Before(x509Certificate.NotAfter)
	if !alive {
		return seederr.WrapErrorf("acme certificate expired at %v", x509Certificate.NotAfter)
	}
	expiringSoon := time.Now().After(x509Certificate.NotAfter.Add(-7 * 24 * time.Hour))
	if expiringSoon {
		return seederr.WrapErrorf("acme certificate expiring soon at %v", x509Certificate.NotAfter)
	}
	return nil
}

type CertificateStorage interface {
	Update(domain string, acmeCertificate *certificate.Resource) error
	Get(domain string) (*certificate.Resource, error)
}

type LocalCertificateStorage struct {
	acmeCertificatesHome string
}

func (s *LocalCertificateStorage) Update(domain string, acmeCertificate *certificate.Resource) error {
	err := checkCertificate(acmeCertificate)
	if err != nil {
		return seederr.Wrap(err)
	}

	sanitizedDomain, err := idna.ToASCII(strings.NewReplacer(":", "-", "*", "_").Replace(domain))
	if err != nil {
		return seederr.Wrap(err)
	}
	acmeCertificateJsonBytes, err := json.Marshal(acmeCertificate, jsontext.WithIndent("  "))
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.MkdirAll(s.acmeCertificatesHome, 0700)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.WriteFile(
		fmt.Sprintf("%s/%s.json", s.acmeCertificatesHome, sanitizedDomain),
		acmeCertificateJsonBytes, 0600)
	if err != nil {
		return seederr.Wrap(err)
	}

	// Also store json ignored fields in corresponding file types.
	err = os.WriteFile(
		fmt.Sprintf("%s/%s.key", s.acmeCertificatesHome, sanitizedDomain),
		acmeCertificate.PrivateKey, 0600)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.WriteFile(
		fmt.Sprintf("%s/%s.crt", s.acmeCertificatesHome, sanitizedDomain),
		acmeCertificate.Certificate, 0600)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.WriteFile(
		fmt.Sprintf("%s/%s.issuer.crt", s.acmeCertificatesHome, sanitizedDomain),
		acmeCertificate.IssuerCertificate, 0600)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.WriteFile(
		fmt.Sprintf("%s/%s.csr", s.acmeCertificatesHome, sanitizedDomain),
		acmeCertificate.CSR, 0600)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func (s *LocalCertificateStorage) Get(domain string) (*certificate.Resource, error) {
	sanitizedDomain, err := idna.ToASCII(strings.NewReplacer(":", "-", "*", "_").Replace(domain))
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	acmeCertificateJsonPath := fmt.Sprintf("%s/%s.json", s.acmeCertificatesHome, sanitizedDomain)
	acmeCertificateJsonBytes, err := os.ReadFile(acmeCertificateJsonPath)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	acmeCertificate := &certificate.Resource{}
	err = json.Unmarshal(acmeCertificateJsonBytes, acmeCertificate)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	acmeCertificate.PrivateKey, err = os.ReadFile(
		fmt.Sprintf("%s/%s.key", s.acmeCertificatesHome, sanitizedDomain))
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	acmeCertificate.Certificate, err = os.ReadFile(
		fmt.Sprintf("%s/%s.crt", s.acmeCertificatesHome, sanitizedDomain))
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	acmeCertificate.IssuerCertificate, err = os.ReadFile(
		fmt.Sprintf("%s/%s.issuer.crt", s.acmeCertificatesHome, sanitizedDomain))
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	acmeCertificate.CSR, err = os.ReadFile(
		fmt.Sprintf("%s/%s.csr", s.acmeCertificatesHome, sanitizedDomain))
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	err = checkCertificate(acmeCertificate)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return acmeCertificate, nil
}

var _ CertificateStorage = &LocalCertificateStorage{}

func NewLocalCertificateStorage(acmeCertificatesHome string) *LocalCertificateStorage {
	return &LocalCertificateStorage{
		acmeCertificatesHome: acmeCertificatesHome,
	}
}
