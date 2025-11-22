package mptauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ndscm/theseed/seed/infra/http/go/seedsession"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type sessionTokenSource struct {
	ctx     context.Context
	prefix  string
	next    oauth2.TokenSource
	session seedsession.SessionAdapter
	last    *oauth2.Token
}

func newSessionTokenSource(
	ctx context.Context,
	prefix string,
	oauth2Config *oauth2.Config,
	session seedsession.SessionAdapter,
	initial *oauth2.Token,
) *sessionTokenSource {
	return &sessionTokenSource{
		ctx:     ctx,
		prefix:  prefix,
		next:    oauth2Config.TokenSource(ctx, initial),
		session: session,
		last:    initial,
	}
}

func (s *sessionTokenSource) Token() (*oauth2.Token, error) {
	newToken, err := s.next.Token()
	if err != nil {
		return nil, err
	}
	if s.last == nil || s.last.AccessToken != newToken.AccessToken {
		if s.session == nil {
			return nil, seederr.WrapErrorf("session is nil")
		}
		err := s.session.Update(s.ctx, map[string]string{
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
	prefix           string
	configurationUrl string
	clientId         string
	clientSecret     string

	cachedConfiguration *OpenidConfiguration
	cachedOrigin        string
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

func (provider *OpenidProvider) getOauth2Config(ctx context.Context, origin string) (*oauth2.Config, error) {
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

func (provider *OpenidProvider) Exchange(
	ctx context.Context,
	session seedsession.SessionAdapter,
	origin string,
	code string,
) error {
	oauth2Config, err := provider.getOauth2Config(ctx, origin)
	if err != nil {
		return seederr.Wrap(err)
	}
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = session.Update(ctx, map[string]string{
		provider.prefix + "access_token":  token.AccessToken,
		provider.prefix + "refresh_token": token.RefreshToken,
		provider.prefix + "expiry":        token.Expiry.Format(time.RFC3339Nano),
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func (provider *OpenidProvider) Client(
	ctx context.Context,
	session seedsession.SessionAdapter,
	origin string,
) (*http.Client, error) {
	oauth2Config, err := provider.getOauth2Config(ctx, origin)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	accessToken, err := session.Get(ctx, provider.prefix+"access_token")
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	refreshToken, err := session.Get(ctx, provider.prefix+"refresh_token")
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	expiryString, err := session.Get(ctx, provider.prefix+"expiry")
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
	tokenSource := newSessionTokenSource(ctx, provider.prefix, oauth2Config, session, initialToken)
	client := oauth2.NewClient(ctx, tokenSource)
	return client, nil
}

func (provider *OpenidProvider) FetchUserInfo(
	ctx context.Context,
	session seedsession.SessionAdapter,
	origin string,
) (*OpenidUserInfo, error) {
	configuration, err := provider.GetOpenidConfiguration(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	client, err := provider.Client(ctx, session, origin)
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
	openidUserInfo := OpenidUserInfo{}
	err = json.Unmarshal(responseBodyBytes, &openidUserInfo)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	err = json.Unmarshal(responseBodyBytes, &openidUserInfo.Raw)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return &openidUserInfo, nil
}

func NewOpenidProvider(prefix string, configurationUrl string, clientId string, clientSecret string) *OpenidProvider {
	return &OpenidProvider{
		prefix:           prefix,
		configurationUrl: configurationUrl,
		clientId:         clientId,
		clientSecret:     clientSecret,
	}
}

type ClientCredentialsOpenidProvider struct {
	OpenidProvider
}

func (provider *ClientCredentialsOpenidProvider) getClientCredentialsConfig(ctx context.Context) (*clientcredentials.Config, error) {
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

func (provider *ClientCredentialsOpenidProvider) Client(
	ctx context.Context,
) (*http.Client, error) {
	oauth2Config, err := provider.getClientCredentialsConfig(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	client := oauth2Config.Client(ctx)
	return client, nil
}

func NewClientCredentialsOpenidProvider(configurationUrl string, clientId string, clientSecret string) *ClientCredentialsOpenidProvider {
	return &ClientCredentialsOpenidProvider{
		OpenidProvider: OpenidProvider{
			configurationUrl: configurationUrl,
			clientId:         clientId,
			clientSecret:     clientSecret,
		},
	}
}
