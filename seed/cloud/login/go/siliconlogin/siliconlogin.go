package siliconlogin

import (
	"context"
	"sync"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"golang.org/x/oauth2"
)

var flagSiliconRefreshToken = seedflag.DefineSecret(
	"silicon_refresh_token",
	"Silicon login refresh token",
)

type SiliconSession struct {
	mutex       sync.Mutex
	provider    *openid.OpenidProvider
	tokenSource oauth2.TokenSource
}

var session = SiliconSession{}

func SiliconLogin(ctx context.Context) (context.Context, error) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	if session.tokenSource == nil || session.provider == nil {
		siliconRefreshToken, err := flagSiliconRefreshToken.LoadString()
		if err != nil {
			return ctx, seederr.Wrap(err)
		}
		if siliconRefreshToken != "" {
			initial := &oauth2.Token{
				RefreshToken: siliconRefreshToken,
			}

			discoveryUrl := openid.OpenidDiscoveryUrlFlag()
			openidClient := openid.NewOpenidClient(discoveryUrl, "silicon-prod", "")
			session.provider = openid.NewOpenidProvider(openidClient, "")
			tokenSource, err := session.provider.WrapExternalTokenStorage(
				ctx, nil, nil, initial,
			)
			if err != nil {
				return ctx, seederr.Wrap(err)
			}
			session.tokenSource = tokenSource
		}
	}

	if session.tokenSource != nil && session.provider != nil {
		token, err := session.tokenSource.Token()
		if err != nil {
			return ctx, seederr.Wrap(err)
		}
		if !token.Valid() {
			return ctx, seederr.WrapErrorf("token is invalid")
		}
		ctx = seedbearer.WithBearer(ctx, token.AccessToken)
	}
	return ctx, nil
}
