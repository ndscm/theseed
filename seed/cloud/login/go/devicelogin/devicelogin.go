package devicelogin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ndscm/theseed/seed/infra/auth/go/loginopenid"
	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/oauth2"
)

var flagLoginTier = seedflag.DefineString("login_tier", "dev", "Login tier (e.g., prod, staging, future, dev)")

type FileTokenStorage struct {
	storagePath string
}

func (s *FileTokenStorage) Get(ctx context.Context, key string) (string, error) {
	mapBytes, err := os.ReadFile(s.storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", seederr.Wrap(err)
	}
	data := map[string]string{}
	err = json.Unmarshal(mapBytes, &data)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return data[key], nil
}

func (s *FileTokenStorage) Update(ctx context.Context, change map[string]string) error {
	data := map[string]string{}
	mapBytes, err := os.ReadFile(s.storagePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return seederr.Wrap(err)
		}
	} else {
		err = json.Unmarshal(mapBytes, &data)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	for k, v := range change {
		data[k] = v
	}
	newMapBytes, err := json.Marshal(data)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.MkdirAll(filepath.Dir(s.storagePath), 0700)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.WriteFile(s.storagePath, newMapBytes, 0600)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

var _ loginopenid.ExternalTokenStorage = &FileTokenStorage{}

func DeviceLogin(ctx context.Context, service string) (context.Context, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return ctx, seederr.Wrap(err)
	}
	discoveryUrl := openid.OpenidDiscoveryUrlFlag()
	serviceTier := service + "-" + flagLoginTier.Get()
	storagePath := filepath.Join(userHome, ".seed", "login", serviceTier+".json")
	tokenStorage := &FileTokenStorage{storagePath: storagePath}

	base := openid.NewOpenidClient(discoveryUrl, serviceTier, "")
	provider := loginopenid.NewUserOpenidProvider(base, "")

	accessToken, err := provider.AccessToken(ctx, tokenStorage)
	if err == nil && accessToken != "" {
		return seedbearer.WithBearer(ctx, accessToken), nil
	}

	oauth2Config, err := base.GetOauth2Config(ctx, "", nil)
	if err != nil {
		return ctx, seederr.Wrap(err)
	}

	verifier := oauth2.GenerateVerifier()
	seedlog.Infof("Requesting device code. endpoint=%s, client_id=%s", oauth2Config.Endpoint.DeviceAuthURL, serviceTier)
	deviceAuth, err := oauth2Config.DeviceAuth(ctx, oauth2.S256ChallengeOption(verifier))
	if err != nil {
		return ctx, seederr.Wrap(err)
	}

	if deviceAuth.VerificationURIComplete != "" {
		fmt.Fprintf(os.Stderr, "To sign in, open: %s\n", deviceAuth.VerificationURIComplete)
	} else {
		fmt.Fprintf(os.Stderr, "To sign in, open %s and enter code: %s\n", deviceAuth.VerificationURI, deviceAuth.UserCode)
	}

	token, err := oauth2Config.DeviceAccessToken(ctx, deviceAuth, oauth2.VerifierOption(verifier))
	if err != nil {
		return ctx, seederr.Wrap(err)
	}

	err = tokenStorage.Update(ctx, map[string]string{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expiry":        token.Expiry.Format(time.RFC3339Nano),
	})
	if err != nil {
		return ctx, seederr.Wrap(err)
	}
	return seedbearer.WithBearer(ctx, token.AccessToken), nil
}
