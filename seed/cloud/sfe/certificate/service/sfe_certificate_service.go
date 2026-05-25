package service

import (
	"context"
	"net/url"
	"strings"

	"connectrpc.com/connect"
	"github.com/go-acme/lego/v5/lego"
	"github.com/ndscm/theseed/seed/cloud/sfe/certificate/account"
	"github.com/ndscm/theseed/seed/cloud/sfe/certificate/challenge"
	"github.com/ndscm/theseed/seed/cloud/sfe/certificate/proto/certificatepb"
	"github.com/ndscm/theseed/seed/cloud/sfe/certificate/storage"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagAcmeAccountEmail = seedflag.DefineString("acme_account_email", "sfe-dev@ndscm.com", "")
var flagAcmeAccountsHome = seedflag.DefineString("acme_accounts_home", "/mnt/data/sfe-certificate/accounts", "")
var flagAcmeCertificatesHome = seedflag.DefineString("acme_certificates_home", "/mnt/data/sfe-certificate/certificates", "")
var flagAcmeProvider = seedflag.DefineString("acme_provider", "letsencrypt", "Provider tag or directory url")

type SfeCertificateChallenger func(ctx context.Context, domain string) (challenge.AcmeChallenge, error)

type SfeCertificateService struct {
	challenger SfeCertificateChallenger
}

func (svc *SfeCertificateService) RenewCertificate(
	ctx context.Context,
	request *connect.Request[certificatepb.RenewCertificateRequest],
) (*connect.Response[certificatepb.RenewCertificateResponse], error) {
	domain := strings.ToLower(strings.Trim(strings.TrimSpace(request.Msg.Domain), "."))
	// The challenger will also check user's permission to request certificate for the domain.
	challengeProvider, err := svc.challenger(ctx, domain)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	certificateStorage := storage.NewLocalCertificateStorage(flagAcmeCertificatesHome.Get())
	acmeCertificate, storageErr := certificateStorage.Get(domain)
	if storageErr != nil {
		seedlog.Infof("Requesting certificate for: %v", domain)
		acmeProvider := flagAcmeProvider.Get()
		if !strings.HasPrefix(acmeProvider, "http") {
			acmeDirectoryUrl, err := lego.GetDirectoryURL(acmeProvider)
			if err != nil {
				return nil, seederr.WrapErrorf("Invalid acme provider %v: %w", acmeProvider, err)
			}
			acmeProvider = acmeDirectoryUrl
		}
		acmeUrl, err := url.Parse(acmeProvider)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		seedlog.Infof("Loaded acme directory. acmeUrl=%v", acmeUrl)

		// Acme Login
		acmeId := strings.NewReplacer(":", "_").Replace(acmeUrl.Host)
		acmeEmail := flagAcmeAccountEmail.Get()
		acmeAccount, err := account.LoadLocalAcmeAccount(flagAcmeAccountsHome.Get(), acmeId, acmeEmail)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		acmeAccountLocker := account.NewLocalAcmeAccountLocker(acmeId, acmeAccount)
		err = acmeAccountLocker.Lock()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		defer func() {
			err := acmeAccountLocker.Unlock()
			if err != nil {
				seedlog.Errorf("%v", err)
			}
		}()

		// Acme Request
		newCertificate, err := challengeProvider.ObtainCertificate(ctx, acmeUrl, acmeAccount, domain)
		if err != nil {
			return nil, seederr.Wrap(err)
		}

		// Store
		err = certificateStorage.Update(domain, newCertificate)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		acmeCertificate = newCertificate
	}
	response := &certificatepb.RenewCertificateResponse{
		Key: acmeCertificate.PrivateKey,
		Crt: acmeCertificate.Certificate,
	}
	return connect.NewResponse(response), nil
}

func NewSfeCertificateService(challenger SfeCertificateChallenger) *SfeCertificateService {
	return &SfeCertificateService{
		challenger: challenger,
	}
}
