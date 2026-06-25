package certstore

import (
	"crypto/tls"
	"crypto/x509"
	"strings"
	"sync"
	"time"

	"github.com/ndscm/theseed/seed/cloud/sfe/certificate/client/go/certificateclient"
	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

func checkCertificateLife(tlsCertificate *tls.Certificate) (bool, bool, error) {
	if len(tlsCertificate.Certificate) == 0 {
		return false, false, seederr.WrapErrorf("empty certificate")
	}
	x509Certificate, err := x509.ParseCertificate(tlsCertificate.Certificate[0])
	if err != nil {
		return false, false, seederr.Wrap(err)
	}
	alive := time.Now().Before(x509Certificate.NotAfter)
	expiringSoon := time.Now().After(x509Certificate.NotAfter.Add(-24 * time.Hour))
	return alive, expiringSoon, nil
}

type SfeCertStore struct {
	serviceOpenid *openid.OpenidClient

	certificateCache sync.Map
}

func (s *SfeCertStore) fetchCertificate(hello *tls.ClientHelloInfo, domain string) (*tls.Certificate, error) {
	ctx := hello.Context()
	client := certificateclient.NewSfeCertificateClient("")
	accessToken, err := s.serviceOpenid.AccessToken(ctx, nil)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	certkey, certBytes, err := client.RenewCertificate(seedbearer.WithBearer(ctx, accessToken), domain)
	if err != nil {
		// Hide the error details
		return nil, seederr.WrapErrorf("invalid domain")
	}
	newCert, err := tls.X509KeyPair(certBytes, certkey)
	if err != nil {
		// Hide the error details
		return nil, seederr.WrapErrorf("invalid certificate")
	}
	s.certificateCache.Store(domain, &newCert)
	return &newCert, nil
}

func (s *SfeCertStore) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	seedlog.Debugf("Tls request: %s", hello.ServerName)
	domain := strings.ToLower(strings.TrimSpace(hello.ServerName))
	if domain == "" {
		return nil, seederr.WrapErrorf("Empty domain in TLS request")
	}
	cert, ok := s.certificateCache.Load(domain)
	if ok {
		certificateInCache := cert.(*tls.Certificate)
		alive, expiringSoon, err := checkCertificateLife(certificateInCache)
		if err != nil {
			seedlog.Warnf("Failed to check certificate life %v: %v", domain, err)
			s.certificateCache.Delete(domain)
		} else if !alive {
			s.certificateCache.Delete(domain)
		} else {
			// alive
			if !expiringSoon {
				return certificateInCache, nil
			}
			newCert, err := s.fetchCertificate(hello, domain)
			if err != nil {
				seedlog.Warnf("Refresh of expiring cert failed for %v: %v", domain, err)
				return certificateInCache, nil
			}
			return newCert, nil
		}
	}
	newCert, err := s.fetchCertificate(hello, domain)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return newCert, nil
}

func NewSfeCertStore(serviceOpenid *openid.OpenidClient) *SfeCertStore {
	return &SfeCertStore{
		serviceOpenid: serviceOpenid,
	}
}
