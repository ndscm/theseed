package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ndscm/theseed/seed/infra/auth/go/openidjwt"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/seedgrpc"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/amadeus/commute/proto/commutepbconnect"
	commuteservice "github.com/ndscm/theseed/seed/newtype/amadeus/commute/service"
	"github.com/ndscm/theseed/seed/newtype/amadeus/onduty"
	"github.com/ndscm/theseed/seed/newtype/amadeus/wake/proto/wakepbconnect"
	wakeservice "github.com/ndscm/theseed/seed/newtype/amadeus/wake/service"
)

var flagPort = seedflag.DefineString("port", "2623", "Server port") // Default port assignment word: AMAD (2623)

// SIGRT3 is SIGRTMIN+3 (signal 37 on Linux). When the container runs with
// podman's --systemd=always, podman uses SIGRTMIN+3 as the stop signal instead
// of SIGTERM, so amadeus-server (PID 1) must handle it to shut down gracefully.
const SIGRT3 = syscall.Signal(37)

func run() error {
	_, err := seedinit.Initialize(
		seedinit.WithEnvPrefix("AMADEUS_"),
		seedinit.WithFallbackEnvPrefix("STEINS_"),
	)
	if err != nil {
		return seederr.Wrap(err)
	}

	conscious, err := onduty.CreateConscious()
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
	err = wakeSvc.Initialize(conscious)
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

	commuteSvc := commuteservice.NewAmadeusCommuteService(conscious)
	err = mux.Register(commutepbconnect.NewAmadeusCommuteServiceHandler(
		commuteSvc,
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}

	handler, err := mux.Ready()
	if err != nil {
		return seederr.Wrap(err)
	}

	conscious.SetConnectHandler(handler)
	err = conscious.Wake()
	if err != nil {
		return seederr.Wrap(err)
	}

	port := flagPort.Get()
	server, err := seedgrpc.CreateServer(":"+port, handler)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer func() {
		const shutdownDrainTimeout = 10 * time.Second

		ctx, cancel := context.WithTimeout(context.Background(), shutdownDrainTimeout)
		defer cancel()
		err := server.Shutdown(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				seedlog.Errorf("Drain timed out after %v, force-closing amadeus server", shutdownDrainTimeout)
				err = server.Close()
			}
		}
		if err != nil {
			seedlog.Errorf("Failed to shutdown amadeus server: %v", err)
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, SIGRT3)
	serveErrChan := make(chan error, 1)
	go func() {
		seedlog.Infof("Starting amadeus server on :%v", port)
		serveErrChan <- server.Serve()
	}()

	select {
	case <-stopChan:
		seedlog.Infof("Shutting down amadeus server")
	case err = <-serveErrChan:
		if err != nil {
			return seederr.Wrap(err)
		}
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
