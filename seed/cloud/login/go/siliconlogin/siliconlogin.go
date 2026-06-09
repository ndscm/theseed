package siliconlogin

import (
	"context"
	"maps"
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

type MemoryTokenStorage struct {
	mutex sync.RWMutex
	data  map[string]string
}

func (s *MemoryTokenStorage) Get(ctx context.Context, key string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if s.data == nil {
		return "", nil
	}
	return s.data[key], nil
}

func (s *MemoryTokenStorage) Update(ctx context.Context, change map[string]string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.data == nil {
		s.data = map[string]string{}
	}
	maps.Copy(s.data, change)
	return nil
}

var _ loginopenid.ExternalTokenStorage = (*MemoryTokenStorage)(nil)

type SiliconSession struct {
	mutex    sync.Mutex
	token    *MemoryTokenStorage
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
			session.token = &MemoryTokenStorage{
				data: map[string]string{"refresh_token": string(refreshToken)},
			}

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
