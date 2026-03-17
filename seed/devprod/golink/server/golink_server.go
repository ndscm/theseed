package main

import (
	"net/http"
	"os"

	"github.com/ndscm/theseed/seed/cloud/login/go/loginservice"
	"github.com/ndscm/theseed/seed/cloud/login/proto/loginpbconnect"
	"github.com/ndscm/theseed/seed/devprod/golink/proto/golinkpbconnect"
	golinkservice "github.com/ndscm/theseed/seed/devprod/golink/service"
	"github.com/ndscm/theseed/seed/infra/http/go/cachecontrol"
	"github.com/ndscm/theseed/seed/infra/spa/go/seedspa"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/grpc/go/seedgrpc"
)

var flagWebapp = seedflag.DefineString("webapp", "", "Path to webapp static files")

func run() error {
	err := seedinit.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}

	mux := &seedgrpc.GrpcMux{}
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

	golinkHandler := &golinkservice.GolinkHandler{}
	if flagWebapp.Get() != "" {
		spaServer := seedspa.SpaServer(http.Dir(flagWebapp.Get()), "__spa-fallback.html")
		spaServer = cachecontrol.InterceptCacheControlMiddleware(spaServer, 3600)
		golinkHandler.Webapp = spaServer
		seedlog.Infof("Serving webapp from: %v", flagWebapp.Get())
	}
	mux.Handle("/", golinkHandler)

	handler, err := mux.Ready()
	if err != nil {
		return seederr.Wrap(err)
	}

	seedlog.Infof("Starting golink server on :4656")
	err = seedgrpc.ListenAndServe(":4656", handler) // GOLN
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
