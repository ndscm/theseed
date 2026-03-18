package clientopenid

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type OpenidConfiguration struct {
	// See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
	Issuer                  string      `json:"issuer"`
	AuthorizationEndpoint   string      `json:"authorization_endpoint"`
	TokenEndpoint           string      `json:"token_endpoint"`
	UserinfoEndpoint        string      `json:"userinfo_endpoint"`
	ScopesSupported         []string    `json:"scopes_supported"`
	ResponsesTypesSupported []string    `json:"responses_types_supported"`
	GrantTypesSupported     []string    `json:"grant_types_supported"`
	SubjectTypesSupported   []string    `json:"subject_types_supported"`
	ClaimsSupported         []string    `json:"claims_supported"`
	Raw                     interface{} `json:"-"`
}

type OpenidUserInfo struct {
	// See: https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
	Sub                 string      `json:"sub"`
	Name                string      `json:"name"`
	GivenName           string      `json:"given_name"`
	FamilyName          string      `json:"family_name"`
	Nickname            string      `json:"nickname"`
	PreferredUsername   string      `json:"preferred_username"`
	Profile             string      `json:"profile"`
	Picture             string      `json:"picture"`
	Website             string      `json:"website"`
	Email               string      `json:"email"`
	EmailVerified       bool        `json:"email_verified"`
	Gender              string      `json:"gender"`
	PhoneNumber         string      `json:"phone_number"`
	PhoneNumberVerified bool        `json:"phone_number_verified"`
	Groups              []string    `json:"groups"`
	Raw                 interface{} `json:"-"`
}

type OpenidProvider struct {
	configurationUrl string
	clientId         string
	clientSecret     string

	cachedConfiguration *OpenidConfiguration
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
	parsedUrl, err := url.Parse(provider.configurationUrl)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	provider.cachedOrigin = parsedUrl.Scheme + "://" + parsedUrl.Host
	return provider.cachedOrigin, nil
}

func (provider *OpenidProvider) GetOpenidConfiguration(ctx context.Context) (*OpenidConfiguration, error) {
	if provider.cachedConfiguration != nil {
		return provider.cachedConfiguration, nil
	}
	seedlog.Infof("Fetching openid configuration from %s", provider.configurationUrl)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.configurationUrl, nil)
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
	provider.cachedConfiguration = &OpenidConfiguration{}
	err = json.Unmarshal(responseBodyBytes, provider.cachedConfiguration)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	err = json.Unmarshal(responseBodyBytes, &provider.cachedConfiguration.Raw)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return provider.cachedConfiguration, nil
}

func (provider *OpenidProvider) GetOauth2Config(ctx context.Context, origin string) (*oauth2.Config, error) {
	configuration, err := provider.GetOpenidConfiguration(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	redirectUri, err := url.Parse(origin)
	if err != nil {
		return nil, seederr.WrapErrorf("invalid origin: %v", err)
	}
	redirectUri.Path = "/auth/callback"
	oauth2Config := &oauth2.Config{
		ClientID:     provider.clientId,
		ClientSecret: provider.clientSecret,
		Scopes:       configuration.ScopesSupported,
		RedirectURL:  redirectUri.String(),
		Endpoint: oauth2.Endpoint{
			AuthURL:  configuration.AuthorizationEndpoint,
			TokenURL: configuration.TokenEndpoint,
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

func (provider *OpenidProvider) Bearer(
	ctx context.Context,
) string {
	if provider.tokenSource == nil {
		oauth2Config, err := provider.GetClientCredentialsConfig(ctx)
		if err != nil {
			return ""
		}
		provider.tokenSource = oauth2Config.TokenSource(context.Background())
	}
	token, err := provider.tokenSource.Token()
	if err != nil {
		return ""
	}
	return "Bearer " + token.AccessToken
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

func NewOpenidProvider(configurationUrl string, clientId string, clientSecret string) *OpenidProvider {
	return &OpenidProvider{
		configurationUrl: configurationUrl,
		clientId:         clientId,
		clientSecret:     clientSecret,
	}
}
