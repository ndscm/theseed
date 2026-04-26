package main

import (
	"os"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/seedgrpc"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/hooin/commute/proto/commutepbconnect"
	commuteservice "github.com/ndscm/theseed/seed/newtype/hooin/commute/service"
	"github.com/ndscm/theseed/seed/newtype/hooin/dictate/proto/dictatepbconnect"
	dictateservice "github.com/ndscm/theseed/seed/newtype/hooin/dictate/service"
)

var flagPort = seedflag.DefineString("port", "4664", "Server port") // Default port assignment word: HOOI (4664)

func run() error {
	_, err := seedinit.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}

	mux := &seedgrpc.GrpcMux{}

	commuteSvc := &commuteservice.HooinCommuteService{}
	err = commuteSvc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = mux.Register(commutepbconnect.NewHooinCommuteServiceHandler(
		commuteSvc,
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}

	dictateSvc := &dictateservice.HooinDictateService{}
	err = dictateSvc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = mux.Register(dictatepbconnect.NewHooinDictateServiceHandler(
		dictateSvc,
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}

	handler, err := mux.Ready()
	if err != nil {
		return seederr.Wrap(err)
	}

	seedlog.Infof("Starting hooin server on :%v", flagPort.Get())
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
