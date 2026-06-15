package siliconlogin

import (
	"context"
	"os"
	"sync"

	"github.com/ndscm/theseed/seed/infra/auth/go/clientopenid"
	"github.com/ndscm/theseed/seed/infra/auth/go/loginopenid"
	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
)

var flagSiliconRefreshTokenFile = seedflag.DefineString("silicon_refresh_token_file", "", "Path to file containing refresh token for Silicon login.")

type SiliconSession struct {
	mutex    sync.Mutex
	token    *loginopenid.MemoryTokenStorage
	provider *loginopenid.UserOpenidProvider
}

var session = SiliconSession{}

func SiliconLogin(ctx context.Context) (context.Context, error) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	if session.token == nil || session.provider == nil {
		siliconRefreshTokenFile := flagSiliconRefreshTokenFile.Get()
		if siliconRefreshTokenFile != "" {
			refreshToken, err := os.ReadFile(siliconRefreshTokenFile)
			if err != nil {
				return ctx, seederr.Wrap(err)
			}
			session.token = loginopenid.NewMemoryTokenStorage(
				map[string]string{"refresh_token": string(refreshToken)},
			)

			discoveryUrl := openid.OpenidDiscoveryUrlFlag()
			base := clientopenid.NewOpenidProvider(discoveryUrl, "silicon-prod", "")
			session.provider = loginopenid.NewUserOpenidProvider(base, "")
		}
	}

	if session.token != nil && session.provider != nil {
		accessToken, err := session.provider.AccessToken(ctx, session.token)
		if err != nil {
			return ctx, seederr.Wrap(err)
		}
		ctx = seedbearer.WithBearer(ctx, accessToken)
	}
	return ctx, nil
}
