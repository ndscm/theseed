package main

import (
	"context"
	"os"
	"slices"

	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/cloud/sfe/certificate/challenge"
	"github.com/ndscm/theseed/seed/cloud/sfe/certificate/challenge/cloudflarednschallenge"
	"github.com/ndscm/theseed/seed/cloud/sfe/certificate/proto/certificatepbconnect"
	certificateservice "github.com/ndscm/theseed/seed/cloud/sfe/certificate/service"
	"github.com/ndscm/theseed/seed/infra/auth/go/openidjwt"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/seedgrpc"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagPort = seedflag.DefineString("port", "7332", "Server port") // Default port assignment word: SFEC (7332)

func challenger(ctx context.Context, domain string) (challenge.AcmeChallenge, error) {
	loginUser, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	seedlog.Infof("Verified login user. user=%v", loginUser)
	switch domain {
	case "kurisu.ndscm.biz":
		permissionDenied := true
		if loginUser.PreferredUsername == "service-account-kurisu-sfe-prod" {
			permissionDenied = false
		}
		if slices.Contains(loginUser.Groups, "sfe-dev") {
			permissionDenied = false
		}
		if permissionDenied {
			return nil, seederr.WrapErrorf("permission denied")
		}
		return cloudflarednschallenge.NewCloudflareDnsChallenge(), nil
	case "workflow.ndscm.biz":
		permissionDenied := true
		if loginUser.PreferredUsername == "service-account-workflow-sfe-prod" {
			permissionDenied = false
		}
		if slices.Contains(loginUser.Groups, "sfe-dev") {
			permissionDenied = false
		}
		if permissionDenied {
			return nil, seederr.WrapErrorf("permission denied")
		}
		return cloudflarednschallenge.NewCloudflareDnsChallenge(), nil
	}
	return nil, seederr.WrapErrorf("invalid domain: %v", domain)
}

func run() error {
	_, err := seedinit.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}

	openidInterceptor, err := openidjwt.CreateOpenidJwtInterceptor()
	if err != nil {
		return seederr.Wrap(err)
	}

	mux, err := seedgrpc.CreateGrpcMux(openidInterceptor.Intercept)
	if err != nil {
		return seederr.Wrap(err)
	}

	certificateSvc := certificateservice.NewSfeCertificateService(challenger)
	err = mux.Register(certificatepbconnect.NewSfeCertificateServiceHandler(
		certificateSvc,
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}

	handler, err := mux.Ready()
	if err != nil {
		return seederr.Wrap(err)
	}

	port := flagPort.Get()
	seedlog.Infof("Starting sfe certificate server. port=%v", port)
	err = seedgrpc.ListenAndServe(":"+port, handler)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		seedlog.Errorf("%v", err)
		os.Exit(1)
	}
}
