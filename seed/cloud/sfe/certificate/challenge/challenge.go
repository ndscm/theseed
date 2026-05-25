package challenge

import (
	"context"
	"net/url"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/registration"
)

type AcmeChallenge interface {
	ObtainCertificate(
		ctx context.Context, acmeDirectoryUrl *url.URL, acmeAccount registration.User, domain string,
	) (*certificate.Resource, error)
}
