package clientopenid

import (
	"context"
	"io"
	"net/http"
	"net/url"

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
