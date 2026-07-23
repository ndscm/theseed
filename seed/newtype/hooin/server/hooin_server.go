package main

import (
	"context"
	"net/http"
	"os"

	"github.com/ndscm/theseed/seed/cloud/bidirequest/go/bidirequest"
	"github.com/ndscm/theseed/seed/cloud/bidirequest/go/bidirequestservice"
	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/auth/go/openidverify"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/ageflag"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/seedgrpc"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/gajetto/team"
	"github.com/ndscm/theseed/seed/newtype/gajetto/team/keycloakteam"
	"github.com/ndscm/theseed/seed/newtype/gajetto/team/staticteam"
	"github.com/ndscm/theseed/seed/newtype/hooin/commute/proto/commutepbconnect"
	commuteservice "github.com/ndscm/theseed/seed/newtype/hooin/commute/service"
	"github.com/ndscm/theseed/seed/newtype/hooin/dictate/proto/dictatepbconnect"
	dictateservice "github.com/ndscm/theseed/seed/newtype/hooin/dictate/service"
	"github.com/ndscm/theseed/seed/newtype/hooin/invade/proto/invadepbconnect"
	invadeservice "github.com/ndscm/theseed/seed/newtype/hooin/invade/service"
	"github.com/ndscm/theseed/seed/newtype/hooin/onsite"
	"github.com/ndscm/theseed/seed/newtype/hooin/raid/proto/raidpbconnect"
	raidservice "github.com/ndscm/theseed/seed/newtype/hooin/raid/service"
	"github.com/ndscm/theseed/seed/newtype/hooin/roster/proto/rosterpbconnect"
	rosterservice "github.com/ndscm/theseed/seed/newtype/hooin/roster/service"
)

var flagPort = seedflag.DefineString(
	"port", "4664", // Default port assignment word: HOOI (4664)
	"Server port",
)
var flagOpenidClientId = seedflag.DefineString(
	"openid_client_id", "",
	"Client ID for OpenID Connect",
)
var flagOpenidClientSecret = ageflag.DefineSecret(
	"openid_client_secret",
	"Client Secret for OpenID Connect",
)

type OfficeConnectHandler struct {
	office *onsite.Office

	personHandler http.Handler
}

func (h *OfficeConnectHandler) HandleConnect(
	ctx context.Context, stream bidirequest.PayloadStream,
) error {

	personId, err := h.office.Team.Auth(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}

	duty := onsite.CreatePersonDuty(stream, h.personHandler)
	err = h.office.SetDuty(personId, duty)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer h.office.ClearDuty(personId)

	<-ctx.Done()

	err = ctx.Err()
	if err != nil && err != context.Canceled {
		return seederr.Wrap(err)
	}
	return nil
}

var _ bidirequestservice.ConnectHandler = (*OfficeConnectHandler)(nil)

func run() error {
	_, err := seedinit.Initialize(
		seedinit.WithFallbackEnvPrefix("SEED_"),
	)
	if err != nil {
		return seederr.Wrap(err)
	}

	t := (team.Team)(nil)
	openidClientId := flagOpenidClientId.Get()
	if openidClientId != "" {
		openidClientSecret, err := flagOpenidClientSecret.LoadString()
		if err != nil {
			return seederr.Wrap(err)
		}
		openidClient := openid.NewOpenidClient(
			openid.OpenidDiscoveryUrlFlag(), openidClientId, openidClientSecret,
		)
		keycloakTeam, err := keycloakteam.ConnectTeam(openidClient)
		if err != nil {
			return seederr.Wrap(err)
		}
		t = keycloakTeam
	} else {
		staticTeam, err := staticteam.LoadTeam()
		if err != nil {
			return seederr.Wrap(err)
		}
		t = staticTeam
	}

	office, err := onsite.CreateOffice(t)
	if err != nil {
		return seederr.Wrap(err)
	}

	openidInterceptor, err := openidverify.CreateOpenidInterceptor()
	if err != nil {
		return seederr.Wrap(err)
	}

	mux, err := seedgrpc.CreateGrpcMux(openidInterceptor.Intercept, seedbearer.InterceptBearerMiddleware)
	if err != nil {
		return seederr.Wrap(err)
	}

	personGrpcMux, err := seedgrpc.CreateGrpcMux(openidInterceptor.Intercept, seedbearer.InterceptBearerMiddleware)
	if err != nil {
		return seederr.Wrap(err)
	}
	commuteSvc := commuteservice.NewHooinCommuteService(office)
	err = personGrpcMux.Register(commutepbconnect.NewHooinCommuteServiceHandler(
		commuteSvc,
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}
	personHandler, err := personGrpcMux.Ready()
	if err != nil {
		return seederr.Wrap(err)
	}

	dictateSvc := dictateservice.NewHooinDictateService(office)
	err = mux.Register(dictatepbconnect.NewHooinDictateServiceHandler(
		dictateSvc,
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}

	invadeSvc := invadeservice.NewHooinInvadeService(office)
	err = mux.Register(invadepbconnect.NewHooinInvadeServiceHandler(
		invadeSvc,
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}

	raidSvc := raidservice.NewHooinRaidService(office)
	err = mux.Register(raidpbconnect.NewHooinRaidServiceHandler(
		raidSvc,
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}

	rosterSvc := rosterservice.NewHooinRosterService(office)
	err = mux.Register(rosterpbconnect.NewHooinRosterServiceHandler(
		rosterSvc,
		seedgrpc.WithCommonInterceptors(),
	))
	if err != nil {
		return seederr.Wrap(err)
	}

	bidirequestPath, bidirequestHandler := bidirequestservice.NewBidirequestServiceHandler(
		&OfficeConnectHandler{
			office:        office,
			personHandler: personHandler,
		},
	)
	bidiHandler := seedbearer.InterceptBearerMiddleware(bidirequestHandler)
	bidiHandler = openidInterceptor.Intercept(bidiHandler)
	mux.Handle(bidirequestPath, bidiHandler)

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
