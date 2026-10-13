package certificatefe

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ndscm/theseed/seed/cloud/sfe/certificate/client/go/certificateclient"
	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/oauth2"
)

func CheckCertificateLife(tlsCertificate *tls.Certificate) (bool, bool, error) {
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

type CertificateStation struct {
	staticBearerToken string

	openidClient *openid.OpenidClient

	openidHttpClient *http.Client

	certificateCache sync.Map
}

func (s *CertificateStation) fetchCertificate(hello *tls.ClientHelloInfo, domain string) (*tls.Certificate, error) {
	ctx := hello.Context()
	client := certificateclient.NewSfeCertificateClient("")
	token := s.staticBearerToken
	if token == "" && s.openidClient != nil {
		tokenCtx := ctx
		if s.openidHttpClient != nil {
			tokenCtx = context.WithValue(ctx, oauth2.HTTPClient, s.openidHttpClient)
		}
		accessToken, err := s.openidClient.AccessToken(tokenCtx, nil)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		token = accessToken
	}
	certkey, certBytes, err := client.RenewCertificate(seedbearer.WithBearer(ctx, token), domain)
	if err != nil {
		// Hide the error details
		return nil, seederr.WrapErrorf("invalid domain: %v", domain)
	}
	newCert, err := tls.X509KeyPair(certBytes, certkey)
	if err != nil {
		// Hide the error details
		return nil, seederr.WrapErrorf("invalid certificate for domain: %v", domain)
	}
	s.certificateCache.Store(domain, &newCert)
	return &newCert, nil
}

func (s *CertificateStation) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	seedlog.Debugf("Tls request: %s", hello.ServerName)
	domain := strings.ToLower(strings.TrimSpace(hello.ServerName))
	if domain == "" {
		return nil, seederr.WrapErrorf("Empty domain in TLS request")
	}
	cert, ok := s.certificateCache.Load(domain)
	if ok {
		certificateInCache := cert.(*tls.Certificate)
		alive, expiringSoon, err := CheckCertificateLife(certificateInCache)
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

type CertificateStationOptions struct {
	StaticBearerToken string

	OpenidClient *openid.OpenidClient

	OpenidHttpClient *http.Client
}

func NewCertificateStation(opts *CertificateStationOptions) *CertificateStation {
	if opts == nil {
		opts = &CertificateStationOptions{}
	}
	return &CertificateStation{
		staticBearerToken: opts.StaticBearerToken,

		openidClient: opts.OpenidClient,

		openidHttpClient: opts.OpenidHttpClient,
	}
}
