package service

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/cloud/keycloak/client/go/keycloak"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/kurisu/proto/kurisupb"
)

type KurisuService struct {
	siliconOpenidProvider *openid.OpenidProvider

	keycloakClient *keycloak.KeycloakClient
}

func (svc *KurisuService) CreateSiliconJwt(
	ctx context.Context,
	req *connect.Request[kurisupb.CreateSiliconJwtRequest],
) (*connect.Response[kurisupb.SiliconJwt], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	personId := req.Msg.GetPersonId()

	// only support keycloak for now

	siliconUuid, err := svc.keycloakClient.GetKeycloakUser(ctx, personId)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	password, err := svc.keycloakClient.GenerateKeycloakUserPassword(ctx, siliconUuid)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	defer func() {
		// There's no fine granted permission to remove a user's credential in keycloak.
		// The removal will fail if the user doesn't have the manage-users permission, but it's anticipated.
		cleanupCtx, cancelCleanup := context.WithTimeout(context.WithoutCancel(ctx), 1*time.Minute)
		defer cancelCleanup()
		err := svc.keycloakClient.RemoveKeycloakUserPassword(cleanupCtx, siliconUuid)
		if err != nil {
			seedlog.Warnf("Silicon user password is kept. personId=%v", personId)
		}
	}()

	tokenSource, err := svc.siliconOpenidProvider.PasswordGrant(
		ctx, personId, password, []string{"openid", "basic", "profile", "email", "offline_access"}, nil,
	)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	token, err := tokenSource.Token()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	refreshToken := token.RefreshToken
	if refreshToken == "" {
		return nil, seederr.WrapErrorf("refresh token is empty")
	}

	jwt := &kurisupb.SiliconJwt{
		RefreshToken: refreshToken,
	}
	return connect.NewResponse(jwt), nil
}

func CreateKurisuService() (*KurisuService, error) {
	siliconOpenidProvider := openid.NewOpenidProvider(
		openid.NewOpenidClient(openid.OpenidDiscoveryUrlFlag(), "silicon-prod", ""),
		"",
	)

	keycloakProvider, err := keycloak.CreateKeycloakClient(
		keycloak.WithDiscoveryUrl(openid.OpenidDiscoveryUrlFlag()),
		keycloak.WithHttpClient(seedbearer.InterceptBearerTransport(http.DefaultClient)),
	)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	svc := &KurisuService{
		siliconOpenidProvider: siliconOpenidProvider,

		keycloakClient: keycloakProvider,
	}
	return svc, nil
}
