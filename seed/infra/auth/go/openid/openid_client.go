package openid

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type OpenidClient struct {
	discoveryUrl string
	clientId     string
	clientSecret string

	cachedConfiguration *OpenidConfiguration
	cachedOrigin        string

	tokenSource oauth2.TokenSource
}

func (oc *OpenidClient) ClientId() string {
	return oc.clientId
}

func (oc *OpenidClient) Origin() (string, error) {
	if oc.cachedOrigin != "" {
		return oc.cachedOrigin, nil
	}
	parsedUrl, err := url.Parse(oc.discoveryUrl)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	oc.cachedOrigin = parsedUrl.Scheme + "://" + parsedUrl.Host
	return oc.cachedOrigin, nil
}

func (oc *OpenidClient) GetOpenidConfiguration(ctx context.Context) (*OpenidConfiguration, error) {
	if oc.cachedConfiguration != nil {
		return oc.cachedConfiguration, nil
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
	oc.cachedConfiguration, err = DecodeOpenidConfiguration(responseBodyBytes)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return oc.cachedConfiguration, nil
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

func (oc *OpenidClient) AccessToken(
	ctx context.Context,
) (string, error) {
	if oc.tokenSource == nil {
		oauth2Config, err := oc.GetClientCredentialsConfig(ctx)
		if err != nil {
			return "", seederr.Wrap(err)
		}
		oc.tokenSource = oauth2Config.TokenSource(context.Background())
	}
	token, err := oc.tokenSource.Token()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return token.AccessToken, nil
}

func (oc *OpenidClient) Client(
	ctx context.Context,
) (*http.Client, error) {
	if oc.tokenSource == nil {
		oauth2Config, err := oc.GetClientCredentialsConfig(ctx)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		oc.tokenSource = oauth2Config.TokenSource(context.Background())
	}
	client := oauth2.NewClient(ctx, oc.tokenSource)
	return client, nil
}

func NewOpenidClient(discoveryUrl string, clientId string, clientSecret string) *OpenidClient {
	return &OpenidClient{
		discoveryUrl: discoveryUrl,
		clientId:     clientId,
		clientSecret: clientSecret,
	}
}
