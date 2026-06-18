package loginopenid

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"golang.org/x/oauth2"
)

type ExternalTokenStorage interface {
	Get(ctx context.Context, key string) (string, error)
	Update(ctx context.Context, change map[string]string) error
}

type externalTokenSource struct {
	ctx     context.Context
	prefix  string
	next    oauth2.TokenSource
	storage ExternalTokenStorage
	last    *oauth2.Token
}

func newExternalTokenSource(
	ctx context.Context,
	prefix string,
	oauth2Config *oauth2.Config,
	storage ExternalTokenStorage,
	initial *oauth2.Token,
) *externalTokenSource {
	return &externalTokenSource{
		ctx:     ctx,
		prefix:  prefix,
		next:    oauth2Config.TokenSource(ctx, initial),
		storage: storage,
		last:    initial,
	}
}

func (s *externalTokenSource) Token() (*oauth2.Token, error) {
	newToken, err := s.next.Token()
	if err != nil {
		return nil, err
	}
	if s.last == nil || s.last.AccessToken != newToken.AccessToken {
		if s.storage == nil {
			return nil, seederr.WrapErrorf("storage is nil")
		}
		err := s.storage.Update(s.ctx, map[string]string{
			s.prefix + "access_token":  newToken.AccessToken,
			s.prefix + "refresh_token": newToken.RefreshToken,
			s.prefix + "expiry":        newToken.Expiry.Format(time.RFC3339Nano),
		})
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		s.last = newToken
	}
	return newToken, nil
}

type UserOpenidProvider struct {
	openid.OpenidClient
	prefix string
}

func (provider *UserOpenidProvider) PasswordGrant(
	ctx context.Context,
	storage ExternalTokenStorage,
	username string,
	password string,
) error {
	oauth2Config, err := provider.GetOauth2Config(
		ctx, "", []string{"openid", "basic", "profile", "email", "offline_access"},
	)
	if err != nil {
		return seederr.Wrap(err)
	}
	token, err := oauth2Config.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = storage.Update(ctx, map[string]string{
		provider.prefix + "access_token":  token.AccessToken,
		provider.prefix + "refresh_token": token.RefreshToken,
		provider.prefix + "expiry":        token.Expiry.Format(time.RFC3339Nano),
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func (provider *UserOpenidProvider) Exchange(
	ctx context.Context,
	storage ExternalTokenStorage,
	origin string,
	code string,
) error {
	oauth2Config, err := provider.GetOauth2Config(ctx, origin, nil)
	if err != nil {
		return seederr.Wrap(err)
	}
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = storage.Update(ctx, map[string]string{
		provider.prefix + "access_token":  token.AccessToken,
		provider.prefix + "refresh_token": token.RefreshToken,
		provider.prefix + "expiry":        token.Expiry.Format(time.RFC3339Nano),
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func (provider *UserOpenidProvider) AccessToken(
	ctx context.Context,
	storage ExternalTokenStorage,
) (string, error) {
	oauth2Config, err := provider.GetOauth2Config(ctx, "", nil)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	accessToken, err := storage.Get(ctx, provider.prefix+"access_token")
	if err != nil {
		return "", seederr.Wrap(err)
	}
	refreshToken, err := storage.Get(ctx, provider.prefix+"refresh_token")
	if err != nil {
		return "", seederr.Wrap(err)
	}
	expiryString, err := storage.Get(ctx, provider.prefix+"expiry")
	if err != nil {
		return "", seederr.Wrap(err)
	}
	expiry, err := time.Parse(time.RFC3339Nano, expiryString)
	if err != nil {
		expiry = time.Now().Add(-time.Minute)
	}
	initialToken := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}
	tokenSource := newExternalTokenSource(ctx, provider.prefix, oauth2Config, storage, initialToken)
	token, err := tokenSource.Token()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return token.AccessToken, nil
}

func (provider *UserOpenidProvider) Authorization(
	ctx context.Context,
	storage ExternalTokenStorage,
) string {
	accessToken, err := provider.AccessToken(ctx, storage)
	if err != nil {
		return ""
	}
	return "Bearer " + accessToken
}

func (provider *UserOpenidProvider) Client(
	ctx context.Context,
	storage ExternalTokenStorage,
	origin string,
) (*http.Client, error) {
	oauth2Config, err := provider.GetOauth2Config(ctx, origin, nil)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	accessToken, err := storage.Get(ctx, provider.prefix+"access_token")
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	refreshToken, err := storage.Get(ctx, provider.prefix+"refresh_token")
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	expiryString, err := storage.Get(ctx, provider.prefix+"expiry")
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if accessToken == "" || refreshToken == "" || expiryString == "" {
		return nil, &oauth2.RetrieveError{ErrorCode: "invalid_grant"}
	}
	expiry, err := time.Parse(time.RFC3339Nano, expiryString)
	if err != nil {
		expiry = time.Now().Add(-time.Minute)
	}
	initialToken := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}
	tokenSource := newExternalTokenSource(ctx, provider.prefix, oauth2Config, storage, initialToken)
	client := oauth2.NewClient(ctx, tokenSource)
	return client, nil
}

func (provider *UserOpenidProvider) FetchUserInfo(
	ctx context.Context,
	storage ExternalTokenStorage,
	origin string,
) (*openid.OpenidUserInfo, error) {
	configuration, err := provider.GetOpenidConfiguration(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	client, err := provider.Client(ctx, storage, origin)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	response, err := client.Get(configuration.UserinfoEndpoint)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	defer response.Body.Close()
	responseBodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, seederr.WrapErrorf("failed to fetch user info: status %d, body: %s",
			response.StatusCode, string(responseBodyBytes))
	}
	openidUserInfo, err := openid.DecodeOpenidUserInfo(responseBodyBytes)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return openidUserInfo, nil
}

func NewUserOpenidProvider(base *openid.OpenidClient, prefix string) *UserOpenidProvider {
	return &UserOpenidProvider{
		OpenidClient: *base,
		prefix:       prefix,
	}
}
