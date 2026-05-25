package main

import (
	"context"
	"os"

	"github.com/ndscm/theseed/seed/cloud/login/go/devicelogin"
	"github.com/ndscm/theseed/seed/cloud/sfe/certificate/client/go/certificateclient"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

func run() error {
	args, err := seedinit.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(args) != 1 {
		return seederr.WrapErrorf("usage: renew <domain>")
	}
	domain := args[0]
	ctx := context.Background()
	ctx, err = devicelogin.DeviceLogin(ctx, "sfe-certificate-device")
	if err != nil {
		return seederr.Wrap(err)
	}
	client := certificateclient.NewSfeCertificateClient("")
	key, crt, err := client.RenewCertificate(ctx, domain)
	if err != nil {
		return seederr.Wrap(err)
	}
	seedlog.Infof("Fetched certificate. domain=%s", domain)
	seedlog.Debugf("Fetched certificate. key=\n%s\ncrt=\n%s\n", string(key), string(crt))
	return nil
}

func main() {
	err := run()
	if err != nil {
		seedlog.Errorf("%v", err)
		os.Exit(1)
	}
}
