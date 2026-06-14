package clientopenid

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type OpenidProvider struct {
	discoveryUrl string
	clientId     string
	clientSecret string

	cachedConfiguration *openid.OpenidConfiguration
	cachedOrigin        string

	tokenSource oauth2.TokenSource
}

func (provider *OpenidProvider) ClientId() string {
	return provider.clientId
}

func (provider *OpenidProvider) Origin() (string, error) {
	if provider.cachedOrigin != "" {
		return provider.cachedOrigin, nil
	}
	parsedUrl, err := url.Parse(provider.discoveryUrl)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	provider.cachedOrigin = parsedUrl.Scheme + "://" + parsedUrl.Host
	return provider.cachedOrigin, nil
}

func (provider *OpenidProvider) GetOpenidConfiguration(ctx context.Context) (*openid.OpenidConfiguration, error) {
	if provider.cachedConfiguration != nil {
		return provider.cachedConfiguration, nil
	}
	seedlog.Infof("Fetching openid configuration from %s", provider.discoveryUrl)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.discoveryUrl, nil)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
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
		return nil, seederr.WrapErrorf("failed to fetch openid configuration: status %d, body: %s",
			response.StatusCode, string(responseBodyBytes))
	}
	provider.cachedConfiguration, err = openid.DecodeOpenidConfiguration(responseBodyBytes)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return provider.cachedConfiguration, nil
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

func (provider *OpenidProvider) GetClientCredentialsConfig(ctx context.Context) (*clientcredentials.Config, error) {
	configuration, err := provider.GetOpenidConfiguration(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	oauth2Config := &clientcredentials.Config{
		ClientID:     provider.clientId,
		ClientSecret: provider.clientSecret,
		TokenURL:     configuration.TokenEndpoint,
		Scopes:       configuration.ScopesSupported,
	}
	return oauth2Config, nil
}

func (provider *OpenidProvider) AccessToken(
	ctx context.Context,
) (string, error) {
	if provider.tokenSource == nil {
		oauth2Config, err := provider.GetClientCredentialsConfig(ctx)
		if err != nil {
			return "", seederr.Wrap(err)
		}
		provider.tokenSource = oauth2Config.TokenSource(context.Background())
	}
	token, err := provider.tokenSource.Token()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return token.AccessToken, nil
}

func (provider *OpenidProvider) Authorization(
	ctx context.Context,
) string {
	accessToken, err := provider.AccessToken(ctx)
	if err != nil {
		return ""
	}
	return "Bearer " + accessToken
}

// TokenExchange performs an RFC 8693 token exchange against the token endpoint,
// authenticating this client with its configured credentials (client_secret_basic
// when a secret is set, otherwise a public client_id).
func (provider *OpenidProvider) TokenExchange(
	ctx context.Context,
	subjectToken string,
	audience string,
	scopes []string,
) (*oauth2.Token, error) {
	configuration, err := provider.GetOpenidConfiguration(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	form.Set("subject_token", subjectToken)
	form.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")
	if audience != "" {
		form.Set("audience", audience)
	}
	if len(scopes) > 0 {
		form.Set("scope", strings.Join(scopes, " "))
	}
	if provider.clientSecret == "" {
		form.Set("client_id", provider.clientId)
	}

	request, err := http.NewRequestWithContext(
		ctx, http.MethodPost, configuration.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if provider.clientSecret != "" {
		request.SetBasicAuth(url.QueryEscape(provider.clientId), url.QueryEscape(provider.clientSecret))
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, seederr.WrapErrorf("token exchange failed: status %d, body: %s",
			response.StatusCode, string(responseBytes))
	}

	offlineToken := &oauth2.Token{}
	err = json.Unmarshal(responseBytes, offlineToken)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	// The wire format carries expires_in, not expiry; derive Expiry so the token
	// is not treated as non-expiring (see oauth2.Token docs).
	if offlineToken.ExpiresIn > 0 {
		offlineToken.Expiry = time.Now().Add(time.Duration(offlineToken.ExpiresIn) * time.Second)
	}
	return offlineToken, nil
}

func (provider *OpenidProvider) Client(
	ctx context.Context,
) (*http.Client, error) {
	if provider.tokenSource == nil {
		oauth2Config, err := provider.GetClientCredentialsConfig(ctx)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		provider.tokenSource = oauth2Config.TokenSource(context.Background())
	}
	client := oauth2.NewClient(ctx, provider.tokenSource)
	return client, nil
}

func NewOpenidProvider(discoveryUrl string, clientId string, clientSecret string) *OpenidProvider {
	return &OpenidProvider{
		discoveryUrl: discoveryUrl,
		clientId:     clientId,
		clientSecret: clientSecret,
	}
}
