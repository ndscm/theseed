package keycloak

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type KeycloakClient struct {
	server string
	realm  string

	httpClient *http.Client
}

// GetKeycloakUser resolves the Keycloak username to the user's
// Keycloak uuid. The request authenticates as the login user via the
// bearer transport, so that user must hold the realm-management view-users role.
func (p *KeycloakClient) GetKeycloakUser(ctx context.Context, username string) (string, error) {
	_, err := uuid.Parse(username)
	if err == nil {
		return username, nil
	}

	query := url.Values{}
	query.Set("username", username)
	query.Set("exact", "true")

	endpoint := p.server + "/admin/realms/" + url.PathEscape(p.realm) + "/users?" + query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	response, err := p.httpClient.Do(request)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	defer response.Body.Close()
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if response.StatusCode != http.StatusOK {
		return "", seederr.WrapErrorf("failed to lookup keycloak user: status %d, body: %s", response.StatusCode, string(responseBytes))
	}

	users := []KeycloakUser{}
	err = json.Unmarshal(responseBytes, &users)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if len(users) == 0 {
		return "", seederr.WrapErrorf("keycloak user not found for personId=%v", username)
	}
	return users[0].Id, nil
}

// GenerateKeycloakUserPassword sets a fresh random password on the user and
// returns it. The password is non-temporary so it can immediately back a
// password grant, and is expected to be removed afterwards.
func (p *KeycloakClient) GenerateKeycloakUserPassword(ctx context.Context, userUuid string) (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	randomPassword := fmt.Sprintf("%x", randomBytes)

	credential := map[string]any{
		"type":      "password",
		"value":     randomPassword,
		"temporary": false,
	}
	credentialBytes, err := json.Marshal(credential)
	if err != nil {
		return "", seederr.Wrap(err)
	}

	// PUT /admin/realms/{realm}/users/{user-id}/reset-password
	endpoint := p.server + "/admin/realms/" + url.PathEscape(p.realm) + "/users/" + url.PathEscape(userUuid) + "/reset-password"
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(credentialBytes))
	if err != nil {
		return "", seederr.Wrap(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := p.httpClient.Do(request)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	defer response.Body.Close()
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if response.StatusCode != http.StatusNoContent {
		return "", seederr.WrapErrorf("failed to set keycloak user password: status %d, body: %s", response.StatusCode, string(responseBytes))
	}
	return randomPassword, nil
}

// RemoveKeycloakUserPassword deletes every password credential on the user so a
// minted temporary password cannot be reused. It is idempotent: removing an
// already-absent password is a no-op.
func (p *KeycloakClient) RemoveKeycloakUserPassword(ctx context.Context, userUuid string) error {
	listEndpoint := p.server + "/admin/realms/" + url.PathEscape(p.realm) + "/users/" + url.PathEscape(userUuid) + "/credentials"
	listRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, listEndpoint, nil)
	if err != nil {
		return seederr.Wrap(err)
	}
	listResponse, err := p.httpClient.Do(listRequest)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer listResponse.Body.Close()
	listBytes, err := io.ReadAll(listResponse.Body)
	if err != nil {
		return seederr.Wrap(err)
	}
	if listResponse.StatusCode != http.StatusOK {
		return seederr.WrapErrorf("failed to list keycloak user credentials: status %d, body: %s", listResponse.StatusCode, string(listBytes))
	}

	credentials := []KeycloakCredential{}
	err = json.Unmarshal(listBytes, &credentials)
	if err != nil {
		return seederr.Wrap(err)
	}

	for _, credential := range credentials {
		if credential.Type != "password" {
			continue
		}
		deleteEndpoint := p.server + "/admin/realms/" + url.PathEscape(p.realm) + "/users/" + url.PathEscape(userUuid) + "/credentials/" + url.PathEscape(credential.Id)
		deleteRequest, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteEndpoint, nil)
		if err != nil {
			return seederr.Wrap(err)
		}
		deleteResponse, err := p.httpClient.Do(deleteRequest)
		if err != nil {
			return seederr.Wrap(err)
		}
		deleteBytes, err := io.ReadAll(deleteResponse.Body)
		deleteResponse.Body.Close()
		if err != nil {
			return seederr.Wrap(err)
		}
		if deleteResponse.StatusCode != http.StatusNoContent {
			return seederr.WrapErrorf("failed to remove keycloak user password: status %d, body: %s", deleteResponse.StatusCode, string(deleteBytes))
		}
	}
	return nil
}

var discoveryPathRegex = regexp.MustCompile("^/realms/([^/]+)/.well-known/openid-configuration$")

type KeycloakClientOptions struct {
	server string
	realm  string

	discoveryUrl string

	httpClient *http.Client
}

func WithServer(server string) func(*KeycloakClientOptions) {
	return func(o *KeycloakClientOptions) {
		o.server = server
	}
}

func WithRealm(realm string) func(*KeycloakClientOptions) {
	return func(o *KeycloakClientOptions) {
		o.realm = realm
	}
}

func WithDiscoveryUrl(discoveryUrl string) func(*KeycloakClientOptions) {
	return func(o *KeycloakClientOptions) {
		o.discoveryUrl = discoveryUrl
	}
}

func WithHttpClient(httpClient *http.Client) func(*KeycloakClientOptions) {
	return func(o *KeycloakClientOptions) {
		o.httpClient = httpClient
	}
}

func CreateKeycloakClient(options ...func(*KeycloakClientOptions)) (*KeycloakClient, error) {
	o := &KeycloakClientOptions{}
	for _, opt := range options {
		opt(o)
	}

	server := o.server
	realm := o.realm
	if o.discoveryUrl != "" {
		discoveryUrl, err := url.Parse(o.discoveryUrl)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		server = discoveryUrl.Scheme + "://" + discoveryUrl.Host
		match := discoveryPathRegex.FindStringSubmatch(discoveryUrl.Path)
		if match == nil {
			return nil, seederr.WrapErrorf("invalid keycloak discovery URL: %v", discoveryUrl)
		}
		realm = url.PathEscape(strings.TrimSpace(match[1]))
	}

	httpClient := o.httpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &KeycloakClient{
		server: server,
		realm:  realm,

		httpClient: httpClient,
	}, nil
}
