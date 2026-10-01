package storage

import (
	"context"
	"time"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func CheckCertificate(acmeCertificate *certificate.Resource) error {
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
	Get(ctx context.Context, domain string) (*certificate.Resource, error)
	Update(ctx context.Context, domain string, acmeCertificate *certificate.Resource) error
}
