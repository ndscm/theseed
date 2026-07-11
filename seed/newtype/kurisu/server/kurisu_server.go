package main

import (
	"embed"
	"io/fs"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/ndscm/theseed/seed/cloud/login/go/loginservice"
	"github.com/ndscm/theseed/seed/cloud/login/proto/loginpbconnect"
	"github.com/ndscm/theseed/seed/devprod/reactrouter/go/reactrouter"
	"github.com/ndscm/theseed/seed/infra/auth/go/openidverify"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/seedgrpc"
	"github.com/ndscm/theseed/seed/infra/http/go/cachecontrol"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/hooin/dictate/client/go/dictateclient"
	"github.com/ndscm/theseed/seed/newtype/hooin/dictate/proto/dictatepbconnect"
	"github.com/ndscm/theseed/seed/newtype/hooin/roster/client/go/rosterclient"
	"github.com/ndscm/theseed/seed/newtype/hooin/roster/proto/rosterpbconnect"
	"github.com/ndscm/theseed/seed/newtype/kurisu/proto/kurisupbconnect"
	kurisuservice "github.com/ndscm/theseed/seed/newtype/kurisu/service"
)

//go:embed all:webapp
var embedFs embed.FS

var flagPort = seedflag.DefineString("port", "5874", "Port") // Default port assignment magic word: KURI

func run() error {
	_, err := seedinit.Initialize(
		seedinit.WithFallbackEnvPrefix("SEED_"),
	)
	if err != nil {
		return seederr.Wrap(err)
	}

	openidInterceptor, err := openidverify.CreateOpenidInterceptor()
	if err != nil {
		return seederr.Wrap(err)
	}

	mux, err := seedgrpc.CreateGrpcMux(
		openidInterceptor.Intercept,
		seedbearer.InterceptBearerMiddleware,
	)
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

	kurisuSvc, err := kurisuservice.CreateKurisuService()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = mux.Register(kurisupbconnect.NewKurisuServiceHandler(
		kurisuSvc,
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}

	dictateUrl, err := url.Parse(dictateclient.HooinDictateServiceServer())
	if err != nil {
		return seederr.Wrap(err)
	}
	dictateProxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(dictateUrl)
			r.SetXForwarded()
		},
	}
	mux.Handle("/"+dictatepbconnect.HooinDictateServiceName+"/", dictateProxy)

	rosterUrl, err := url.Parse(rosterclient.HooinRosterServiceServer())
	if err != nil {
		return seederr.Wrap(err)
	}
	rosterProxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(rosterUrl)
			r.SetXForwarded()
		},
	}
	mux.Handle("/"+rosterpbconnect.HooinRosterServiceName+"/", rosterProxy)

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

	seedlog.Infof("Starting kurisu server on :%v", flagPort.Get())
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
