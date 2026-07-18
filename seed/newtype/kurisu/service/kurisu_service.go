package service

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/ndscm/theseed/seed/cloud/keycloak/client/go/keycloak"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
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

	siliconUser := (*keycloak.UserRepresentation)(nil)
	_, err = uuid.Parse(personId)
	if err != nil {
		// guess the personId is actually the person handle
		siliconUser, err = svc.keycloakClient.GetUserByHandle(ctx, true, personId)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	}
	siliconUuid := siliconUser.Id
	siliconUsername := siliconUser.Username

	password, err := svc.keycloakClient.GenerateUserPassword(ctx, true, siliconUuid)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	defer func() {
		// There's no fine granted permission to remove a user's credential in keycloak.
		// The removal will fail if the user doesn't have the manage-users permission, but it's anticipated.
		cleanupCtx, cancelCleanup := context.WithTimeout(context.WithoutCancel(ctx), 1*time.Minute)
		defer cancelCleanup()
		err := svc.keycloakClient.RemoveUserPassword(cleanupCtx, true, siliconUuid)
		if err != nil {
			seedlog.Warnf("Silicon user password is kept. personId=%v", personId)
		}
	}()

	tokenSource, err := svc.siliconOpenidProvider.PasswordGrant(
		ctx, siliconUsername, password, []string{"openid", "basic", "profile", "email", "offline_access"}, nil,
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

	kurisuOpenidClient := openid.NewOpenidClient(openid.OpenidDiscoveryUrlFlag(), "kurisu-prod", "")
	keycloakProvider, err := keycloak.CreateKeycloakClient(kurisuOpenidClient)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	svc := &KurisuService{
		siliconOpenidProvider: siliconOpenidProvider,

		keycloakClient: keycloakProvider,
	}
	return svc, nil
}
