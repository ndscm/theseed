package openid

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// clientAssertionTransport authenticates the OpenID client to the token
// endpoint using a JWT client assertion (RFC 7523) instead of a client secret.
// The assertion is the access token fetched from tokenSource (e.g. SFE's own
// access token), injected into form-encoded token requests so it covers both
// the initial code exchange and subsequent refreshes.
type clientAssertionTransport struct {
	next http.RoundTripper

	clientAssertion oauth2.TokenSource
}

func (t *clientAssertionTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.Method != http.MethodPost ||
		request.Header.Get("Content-Type") != "application/x-www-form-urlencoded" ||
		request.Body == nil {
		return t.next.RoundTrip(request)
	}

	token, err := t.clientAssertion.Token()
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	err = request.Body.Close()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	form, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	// The client is identified entirely by the assertion (iss/sub), so drop any
	// client_id/client_secret that oauth2 auto-appends. Keycloak skips its
	// client_id consistency check when client_id is absent.
	form.Del("client_id")
	form.Del("client_secret")

	form.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	form.Set("client_assertion", token.AccessToken)

	encodedForm := form.Encode()
	clonedRequest := request.Clone(request.Context())
	clonedRequest.Body = io.NopCloser(strings.NewReader(encodedForm))
	clonedRequest.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(encodedForm)), nil
	}
	clonedRequest.ContentLength = int64(len(encodedForm))
	return t.next.RoundTrip(clonedRequest)
}

var _ http.RoundTripper = (*clientAssertionTransport)(nil)

type OpenidClient struct {
	discoveryUrl string
	clientId     string
	clientSecret string

	clientAssertion oauth2.TokenSource

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
		Scopes:       []string{"openid"},
	}
	return oauth2Config, nil
}

func (oc *OpenidClient) WithClientAssertion(ctx context.Context) context.Context {
	if oc.clientAssertion == nil {
		return ctx
	}
	next := http.DefaultTransport
	originalClient, ok := ctx.Value(oauth2.HTTPClient).(*http.Client)
	if ok && originalClient != nil && originalClient.Transport != nil {
		next = originalClient.Transport
	}
	newClient := &http.Client{
		Transport: &clientAssertionTransport{
			next: next,

			clientAssertion: oc.clientAssertion,
		},
	}
	return context.WithValue(ctx, oauth2.HTTPClient, newClient)
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
	refreshCtx := oc.WithClientAssertion(context.Background())
	tokenSource := oauth2Config.TokenSource(refreshCtx)
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

func NewOpenidClientWithAssertion(discoveryUrl string, clientId string, tokenSource oauth2.TokenSource) *OpenidClient {
	return &OpenidClient{
		discoveryUrl:    discoveryUrl,
		clientId:        clientId,
		clientAssertion: tokenSource,
	}
}
