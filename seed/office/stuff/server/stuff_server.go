package main

import (
	"net/http"
	"os"

	"github.com/ndscm/theseed/seed/cloud/login/go/loginservice"
	"github.com/ndscm/theseed/seed/cloud/login/proto/loginpbconnect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/seedgrpc"
	"github.com/ndscm/theseed/seed/infra/http/go/cachecontrol"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/spa/go/seedspa"
	"github.com/ndscm/theseed/seed/office/stuff/proto/stuffpbconnect"
	stuffservice "github.com/ndscm/theseed/seed/office/stuff/service"
)

var flagWebapp = seedflag.DefineString("webapp", "", "Path to webapp static files")
var flagPort = seedflag.DefineString("port", "7883", "Port to run the server on") // STUF

func run() error {
	_, err := seedinit.Initialize()
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
	err = mux.Register(stuffpbconnect.NewStuffServiceHandler(
		&stuffservice.StuffService{},
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}

	if flagWebapp.Get() != "" {
		spaServer := seedspa.SpaServer(http.Dir(flagWebapp.Get()), "__spa-fallback.html")
		spaServer = cachecontrol.InterceptCacheControlMiddleware(spaServer, 3600)
		mux.Handle("/", spaServer)
		seedlog.Infof("Serving webapp from: %v", flagWebapp.Get())
	}

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
