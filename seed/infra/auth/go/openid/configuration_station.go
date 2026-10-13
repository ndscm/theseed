package openid

import (
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/oauth2"
	"golang.org/x/sync/singleflight"
)

type openidConfigurationStation struct {
	mutex sync.RWMutex

	configurationMap map[string]*OpenidConfiguration

	fetchGroup singleflight.Group
}

func (s *openidConfigurationStation) get(discoveryUrl string) (*OpenidConfiguration, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	configuration, ok := s.configurationMap[discoveryUrl]
	return configuration, ok
}

func (s *openidConfigurationStation) set(discoveryUrl string, configuration *OpenidConfiguration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.configurationMap[discoveryUrl] = configuration
}

func (s *openidConfigurationStation) request(ctx context.Context, discoveryUrl string) (*OpenidConfiguration, error) {
	seedlog.Infof("Fetching openid configuration from %s", discoveryUrl)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryUrl, nil)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	httpClient, ok := ctx.Value(oauth2.HTTPClient).(*http.Client)
	if !ok || httpClient == nil {
		httpClient = http.DefaultClient
	}

	response, err := httpClient.Do(request)
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

	s.set(discoveryUrl, configuration)
	return configuration, nil
}

func (s *openidConfigurationStation) fetch(ctx context.Context, discoveryUrl string) (*OpenidConfiguration, error) {
	cached, ok := s.get(discoveryUrl)
	if ok && cached != nil {
		return cached, nil
	}

	// Collapse a burst of concurrent cold requests for the same URL into one
	// outbound fetch; the rest wait and share its configuration.
	v, err, _ := s.fetchGroup.Do(discoveryUrl,
		func() (any, error) {
			cached, ok := s.get(discoveryUrl)
			if ok && cached != nil {
				return cached, nil
			}
			return s.request(ctx, discoveryUrl)
		},
	)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	configuration := v.(*OpenidConfiguration)
	return configuration, nil
}

var singletonOpenidConfigurationStation = &openidConfigurationStation{
	configurationMap: map[string]*OpenidConfiguration{},
}

func FetchOpenidConfiguration(ctx context.Context, discoveryUrl string) (*OpenidConfiguration, error) {
	return singletonOpenidConfigurationStation.fetch(ctx, discoveryUrl)
}
