package main

import (
	"context"
	"embed"
	"io/fs"
	"os"
	"time"

	"github.com/ndscm/theseed/seed/cloud/login/go/loginservice"
	"github.com/ndscm/theseed/seed/cloud/login/proto/loginpbconnect"
	"github.com/ndscm/theseed/seed/devprod/reactrouter/go/reactrouter"
	"github.com/ndscm/theseed/seed/infra/auth/go/openidverify"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/seedgrpc"
	"github.com/ndscm/theseed/seed/infra/http/go/cachecontrol"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/office/stuff/database/stuffdb"
	"github.com/ndscm/theseed/seed/office/stuff/proto/stuffpbconnect"
	stuffservice "github.com/ndscm/theseed/seed/office/stuff/service"
)

//go:embed all:webapp
var embedFs embed.FS

var flagPort = seedflag.DefineString("port", "7883", "Port") // Default port assignment magic word: STUF

func run() error {
	_, err := seedinit.Initialize(
		seedinit.WithFallbackEnvPrefix("SEED_"),
	)
	if err != nil {
		return seederr.Wrap(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := stuffdb.Open(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = db.Schema.Create(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}

	openidInterceptor, err := openidverify.CreateOpenidInterceptor()
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
	err = mux.Register(stuffpbconnect.NewStuffServiceHandler(
		&stuffservice.StuffService{},
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}

	webapp, err := fs.Sub(embedFs, "webapp")
	if err != nil {
		return seederr.Wrap(err)
	}
	spaServer, _, err := reactrouter.I18nSpaServer(webapp)
	if err != nil {
		return seederr.Wrap(err)
	}
	spaServer = cachecontrol.InterceptCacheControlMiddleware(spaServer, 3600)
	mux.Handle("/", spaServer)

	handler, err := mux.Ready()
	if err != nil {
		return seederr.Wrap(err)
	}

	seedlog.Infof("Starting stuff server on :%v", flagPort.Get())
	err = seedgrpc.ListenAndServe(":"+flagPort.Get(), handler)
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
