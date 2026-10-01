package localstorage

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"os"
	"strings"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/ndscm/theseed/seed/cloud/sfe/certificate/storage"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"golang.org/x/net/idna"
)

var flagAcmeCertificatesHome = seedflag.DefineString(
	"acme_certificates_home", "/mnt/data/sfe-certificate/certificates",
	"",
)

type LocalCertificateStorage struct {
	acmeCertificatesHome string
}

func (s *LocalCertificateStorage) Get(ctx context.Context, domain string) (*certificate.Resource, error) {
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

	err = storage.CheckCertificate(acmeCertificate)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return acmeCertificate, nil
}

func (s *LocalCertificateStorage) Update(ctx context.Context, domain string, acmeCertificate *certificate.Resource) error {
	err := storage.CheckCertificate(acmeCertificate)
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

var _ storage.CertificateStorage = &LocalCertificateStorage{}

func NewLocalCertificateStorage(acmeCertificatesHome string) *LocalCertificateStorage {
	return &LocalCertificateStorage{
		acmeCertificatesHome: acmeCertificatesHome,
	}
}

func LoadLocalCertificateStorage() *LocalCertificateStorage {
	return NewLocalCertificateStorage(flagAcmeCertificatesHome.Get())
}
