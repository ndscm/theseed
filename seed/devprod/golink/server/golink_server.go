package main

import (
	"embed"
	"io/fs"
	"os"

	"github.com/ndscm/theseed/seed/cloud/login/go/loginservice"
	"github.com/ndscm/theseed/seed/cloud/login/proto/loginpbconnect"
	"github.com/ndscm/theseed/seed/devprod/golink/proto/golinkpbconnect"
	golinkservice "github.com/ndscm/theseed/seed/devprod/golink/service"
	"github.com/ndscm/theseed/seed/devprod/reactrouter/go/reactrouter"
	"github.com/ndscm/theseed/seed/infra/auth/go/openidjwt"
	"github.com/ndscm/theseed/seed/infra/buildinfo/go/buildinfo"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/seedgrpc"
	"github.com/ndscm/theseed/seed/infra/http/go/cachecontrol"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/spa/go/seedspa"
)

//go:embed all:webapp
var embedFs embed.FS

var flagPort = seedflag.DefineString("port", "4656", "Port") // Default port assignment magic word: GOLN

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

	err = mux.Register(loginpbconnect.NewLoginServiceHandler(
		&loginservice.LoginService{},
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}
	err = mux.Register(golinkpbconnect.NewGolinkServiceHandler(
		&golinkservice.GolinkService{},
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}

	webapp, err := fs.Sub(embedFs, "webapp")
	if err != nil {
		return seederr.Wrap(err)
	}
	buildinfoInjection, err := buildinfo.GenerateWebappHeadInjection()
	if err != nil {
		return seederr.Wrap(err)
	}
	spaServer, extraLanguages, err := reactrouter.I18nSpaServer(webapp, seedspa.WithHeadInjection(buildinfoInjection))
	if err != nil {
		return seederr.Wrap(err)
	}
	spaServer = cachecontrol.InterceptCacheControlMiddleware(spaServer, 3600)
	golinkHandler := golinkservice.NewGolinkHandler(spaServer, extraLanguages)
	mux.Handle("/", golinkHandler)

	handler, err := mux.Ready()
	if err != nil {
		return seederr.Wrap(err)
	}

	port := flagPort.Get()
	seedlog.Infof("Starting golink server. port=%v", port)
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
