package openid

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"golang.org/x/oauth2"
)

type OpenidProvider struct {
	*OpenidClient
	prefix string
}

// See: https://datatracker.ietf.org/doc/html/rfc6749#section-4.1
func (provider *OpenidProvider) CodeGrant(
	ctx context.Context,
	origin string,
	code string,
	scopes []string,
	storage ExternalTokenStorage,
) (oauth2.TokenSource, error) {
	oauth2Config, err := provider.GetOauth2Config(ctx, origin, scopes)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if storage == nil {
		return oauth2Config.TokenSource(context.Background(), token), nil
	}
	userTokenSource, err := createExternalTokenStorageTokenSource(
		context.Background(), provider.prefix, oauth2Config, storage, token,
	)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return userTokenSource, nil
}

// See: https://datatracker.ietf.org/doc/html/rfc6749#section-4.3
func (provider *OpenidProvider) PasswordGrant(
	ctx context.Context,
	username string,
	password string,
	scopes []string,
	storage ExternalTokenStorage,
) (oauth2.TokenSource, error) {
	oauth2Config, err := provider.GetOauth2Config(ctx, "", scopes)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	token, err := oauth2Config.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if storage == nil {
		return oauth2Config.TokenSource(context.Background(), token), nil
	}
	userTokenSource, err := createExternalTokenStorageTokenSource(
		context.Background(), provider.prefix, oauth2Config, storage, token,
	)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return userTokenSource, nil
}

func (provider *OpenidProvider) AccessToken(
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
	tokenSource, err := createExternalTokenStorageTokenSource(
		ctx, provider.prefix, oauth2Config, storage, initialToken,
	)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	token, err := tokenSource.Token()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return token.AccessToken, nil
}

func (provider *OpenidProvider) Authorization(
	ctx context.Context,
	storage ExternalTokenStorage,
) string {
	accessToken, err := provider.AccessToken(ctx, storage)
	if err != nil {
		return ""
	}
	return "Bearer " + accessToken
}

func (provider *OpenidProvider) Client(
	ctx context.Context,
	storage ExternalTokenStorage,
	origin string,
) (*http.Client, error) {
	oauth2Config, err := provider.GetOauth2Config(ctx, origin, nil)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	tokenSource, err := createExternalTokenStorageTokenSource(
		ctx, provider.prefix, oauth2Config, storage, nil,
	)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	token, err := tokenSource.Token()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if !token.Valid() {
		return nil, seederr.WrapErrorf("invalid token")
	}
	client := oauth2.NewClient(ctx, tokenSource)
	return client, nil
}

func (provider *OpenidProvider) FetchUserInfo(
	ctx context.Context,
	storage ExternalTokenStorage,
	origin string,
) (*OpenidUserInfo, error) {
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
	openidUserInfo, err := DecodeOpenidUserInfo(responseBodyBytes)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return openidUserInfo, nil
}

func NewOpenidProvider(openidClient *OpenidClient, prefix string) *OpenidProvider {
	return &OpenidProvider{
		OpenidClient: openidClient,
		prefix:       prefix,
	}
}
