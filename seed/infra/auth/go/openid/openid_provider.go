package openid

import (
	"context"
	"encoding/json/v2"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"golang.org/x/oauth2"
)

type OpenidProvider struct {
	*OpenidClient
	prefix string
}

func (provider *OpenidProvider) GetOauth2Config(
	ctx context.Context, origin string, scopes []string,
) (*oauth2.Config, error) {
	configuration, err := provider.GetOpenidConfiguration(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	redirectUrl := ""
	if origin != "" {
		redirectUri, err := url.Parse(origin)
		if err != nil {
			return nil, seederr.WrapErrorf("invalid origin: %v", err)
		}
		redirectUri.Path = "/auth/callback"
		redirectUrl = redirectUri.String()
	}

	authStyle := oauth2.AuthStyleAutoDetect
	if provider.clientSecret == "" {
		authStyle = oauth2.AuthStyleInParams
	}

	if len(scopes) == 0 {
		scopes = configuration.ScopesSupported
	}

	oauth2Config := &oauth2.Config{
		ClientID:     provider.clientId,
		ClientSecret: provider.clientSecret,
		Scopes:       scopes,
		RedirectURL:  redirectUrl,
		Endpoint: oauth2.Endpoint{
			AuthURL:       configuration.AuthorizationEndpoint,
			DeviceAuthURL: configuration.DeviceAuthorizationEndpoint,
			TokenURL:      configuration.TokenEndpoint,
			AuthStyle:     authStyle,
		},
	}
	return oauth2Config, nil
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
	ctx = provider.WithClientAssertion(ctx)
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	refreshCtx := provider.WithClientAssertion(context.Background())
	if storage == nil {
		return oauth2Config.TokenSource(refreshCtx, token), nil
	}
	userTokenSource, err := createExternalTokenStorageTokenSource(
		refreshCtx, provider.prefix, oauth2Config, storage, token,
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
	ctx = provider.WithClientAssertion(ctx)
	token, err := oauth2Config.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	refreshCtx := provider.WithClientAssertion(context.Background())
	if storage == nil {
		return oauth2Config.TokenSource(refreshCtx, token), nil
	}
	userTokenSource, err := createExternalTokenStorageTokenSource(
		refreshCtx, provider.prefix, oauth2Config, storage, token,
	)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return userTokenSource, nil
}

func (provider *OpenidProvider) WrapExternalTokenStorage(
	ctx context.Context,
	scopes []string,
	storage ExternalTokenStorage,
	initial *oauth2.Token,
) (oauth2.TokenSource, error) {
	oauth2Config, err := provider.GetOauth2Config(ctx, "", scopes)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	refreshCtx := provider.WithClientAssertion(context.Background())
	if storage == nil {
		return oauth2Config.TokenSource(refreshCtx, initial), nil
	}
	userTokenSource, err := createExternalTokenStorageTokenSource(
		refreshCtx, provider.prefix, oauth2Config, storage, initial,
	)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return userTokenSource, nil
}

// UmaGrant exchanges a subject's access token for a Requesting Party Token (RPT)
// via the UMA ticket grant, persists the RPT to storage, and returns a
// TokenSource that refreshes it through the standard refresh_token grant. The
// RPT carries the authorization claim, so services need not fetch it
// themselves. Keycloak hands back another RPT on refresh, so the stored bearer
// keeps its claim without re-running the exchange.
//
// The exchange requests no particular permission, so Keycloak returns every
// grant the subject holds on the audience's resources. It authenticates as the
// subject through the access token in the Authorization header, not as the
// client.
func (provider *OpenidProvider) UmaGrant(
	ctx context.Context,
	subjectToken string,
	audience string,
	storage ExternalTokenStorage,
) (oauth2.TokenSource, error) {
	configuration, err := provider.GetOpenidConfiguration(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:uma-ticket")
	form.Set("audience", audience)

	request, err := http.NewRequestWithContext(
		ctx, http.MethodPost, configuration.TokenEndpoint, strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Bearer "+subjectToken)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	defer response.Body.Close()
	responseBodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, seederr.WrapErrorf("uma ticket grant failed: status %d, body: %s",
			response.StatusCode, string(responseBodyBytes))
	}

	// oauth2.Token carries the token endpoint's wire tags (access_token,
	// refresh_token, expires_in, ...), so the reply decodes straight into it.
	// Populating Expiry from the relative expires_in is left to the caller.
	token := &oauth2.Token{}
	err = json.Unmarshal(responseBodyBytes, token)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if token.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}

	tokenSource, err := provider.WrapExternalTokenStorage(ctx, nil, storage, token)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return tokenSource, nil
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
	refreshCtx := provider.WithClientAssertion(context.Background())
	tokenSource, err := createExternalTokenStorageTokenSource(
		refreshCtx, provider.prefix, oauth2Config, storage, nil,
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
