package main

import (
	"os"

	"github.com/ndscm/theseed/seed/infra/auth/go/openidjwt"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/seedgrpc"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/amadeus/wake/proto/wakepbconnect"
	wakeservice "github.com/ndscm/theseed/seed/newtype/amadeus/wake/service"
)

var flagPort = seedflag.DefineString("port", "2623", "Server port") // Default port assignment word: AMAD (2623)

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

	wakeSvc := &wakeservice.AmadeusWakeService{}
	err = wakeSvc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = mux.Register(wakepbconnect.NewAmadeusWakeServiceHandler(
		wakeSvc,
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}

	handler, err := mux.Ready()
	if err != nil {
		return seederr.Wrap(err)
	}

	seedlog.Infof("Starting amadeus server on :%v", flagPort.Get())
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
