package cloudflarednschallenge

import (
	"context"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/providers/dns/cloudflare"
	"github.com/go-acme/lego/v5/registration"
	"github.com/ndscm/theseed/seed/cloud/sfe/certificate/challenge"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagCloudflareDnsApiTokenFile = seedflag.DefineString("cloudflare_dns_api_token_file", "user-secret://sfe-certificate/CLOUDFLARE_DNS_API_TOKEN/latest", "Cloudflare dns api token file path")

type CloudflareDnsChallenge struct {
}

func (c *CloudflareDnsChallenge) ObtainCertificate(
	ctx context.Context, acmeDirectoryUrl *url.URL, acmeAccount registration.User, domain string,
) (*certificate.Resource, error) {
	legoConfig := lego.NewConfig(acmeAccount)
	legoConfig.CADirURL = acmeDirectoryUrl.String()
	legoClient, err := lego.NewClient(legoConfig)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	tokenBytes, err := os.ReadFile(flagCloudflareDnsApiTokenFile.Get())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	cloudflareConfig := cloudflare.NewDefaultConfig()
	cloudflareConfig.AuthToken = strings.TrimSpace(string(tokenBytes))
	cloudflareDnsProvider, err := cloudflare.NewDNSProviderConfig(cloudflareConfig)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	dns01Client := dns01.NewClient(&dns01.Options{
		RecursiveNameservers: []string{"1.1.1.1:53"},
	})
	dns01.SetDefaultClient(dns01Client)
	err = legoClient.Challenge.SetDNS01Provider(cloudflareDnsProvider)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	// Keep below fc's 300s timeout so cleanup can run before forced termination.
	ctx, cancel := context.WithTimeout(ctx, 240*time.Second)
	defer cancel()
	acmeCertificate, err := legoClient.Certificate.Obtain(ctx, certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
		KeyType: certcrypto.RSA3072,
	})
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return acmeCertificate, nil
}

func NewCloudflareDnsChallenge() challenge.AcmeChallenge {
	return &CloudflareDnsChallenge{}
}
