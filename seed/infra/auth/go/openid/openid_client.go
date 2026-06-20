package openid

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type OpenidClient struct {
	discoveryUrl string
	clientId     string
	clientSecret string

	configurationMutex sync.RWMutex
	configuration      *OpenidConfiguration

	tokenSourceMutex sync.RWMutex
	tokenSource      oauth2.TokenSource
}

func (oc *OpenidClient) ClientId() string {
	return oc.clientId
}

func (oc *OpenidClient) Origin() (string, error) {
	parsedUrl, err := url.Parse(oc.discoveryUrl)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	origin := parsedUrl.Scheme + "://" + parsedUrl.Host
	return origin, nil
}

func (oc *OpenidClient) GetOpenidConfiguration(ctx context.Context) (*OpenidConfiguration, error) {
	cached := func() *OpenidConfiguration {
		oc.configurationMutex.RLock()
		defer oc.configurationMutex.RUnlock()
		return oc.configuration
	}()
	if cached != nil {
		return cached, nil
	}
	seedlog.Infof("Fetching openid configuration from %s", oc.discoveryUrl)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, oc.discoveryUrl, nil)
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
	configuration, err := DecodeOpenidConfiguration(responseBodyBytes)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	func() {
		oc.configurationMutex.Lock()
		defer oc.configurationMutex.Unlock()
		oc.configuration = configuration
	}()
	return configuration, nil
}

func (oc *OpenidClient) GetClientCredentialsConfig(ctx context.Context) (*clientcredentials.Config, error) {
	configuration, err := oc.GetOpenidConfiguration(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	oauth2Config := &clientcredentials.Config{
		ClientID:     oc.clientId,
		ClientSecret: oc.clientSecret,
		TokenURL:     configuration.TokenEndpoint,
		Scopes:       configuration.ScopesSupported,
	}
	return oauth2Config, nil
}

func (oc *OpenidClient) GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	cached := func() oauth2.TokenSource {
		oc.tokenSourceMutex.RLock()
		defer oc.tokenSourceMutex.RUnlock()
		return oc.tokenSource
	}()
	if cached != nil {
		return cached, nil
	}
	oauth2Config, err := oc.GetClientCredentialsConfig(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	tokenSource := oauth2Config.TokenSource(context.Background())
	func() {
		oc.tokenSourceMutex.Lock()
		defer oc.tokenSourceMutex.Unlock()
		oc.tokenSource = tokenSource
	}()
	return tokenSource, nil
}

func (oc *OpenidClient) AccessToken(
	ctx context.Context,
) (string, error) {
	tokenSource, err := oc.GetTokenSource(ctx)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	token, err := tokenSource.Token()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return token.AccessToken, nil
}

func (oc *OpenidClient) Client(
	ctx context.Context,
) (*http.Client, error) {
	tokenSource, err := oc.GetTokenSource(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	client := oauth2.NewClient(ctx, tokenSource)
	return client, nil
}

func NewOpenidClient(discoveryUrl string, clientId string, clientSecret string) *OpenidClient {
	return &OpenidClient{
		discoveryUrl: discoveryUrl,
		clientId:     clientId,
		clientSecret: clientSecret,
	}
}
