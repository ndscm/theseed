package siliconlogin

import (
	"context"
	"os"
	"sync"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"golang.org/x/oauth2"
)

var flagSiliconRefreshTokenFile = seedflag.DefineString("silicon_refresh_token_file", "", "Path to file containing refresh token for Silicon login.")

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
		siliconRefreshTokenFile := flagSiliconRefreshTokenFile.Get()
		if siliconRefreshTokenFile != "" {
			refreshToken, err := os.ReadFile(siliconRefreshTokenFile)
			if err != nil {
				return ctx, seederr.Wrap(err)
			}
			initial := &oauth2.Token{
				RefreshToken: string(refreshToken),
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
