package main

import (
	"os"

	"github.com/ndscm/theseed/seed/cloud/login/go/loginservice"
	"github.com/ndscm/theseed/seed/cloud/login/proto/loginpbconnect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/grpc/go/seedgrpc"
	"github.com/ndscm/theseed/seed/office/stuff/proto/stuffpbconnect"
	stuffservice "github.com/ndscm/theseed/seed/office/stuff/service"
)

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
	err = mux.Register(stuffpbconnect.NewStuffServiceHandler(
		&stuffservice.StuffService{},
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}

	handler, err := mux.Ready()
	if err != nil {
		return seederr.Wrap(err)
	}

	seedlog.Infof("Starting stuff server on :7883")
	err = seedgrpc.ListenAndServe(":7883", handler) // STUF
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
