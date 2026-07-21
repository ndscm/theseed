package keycloak

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json/v2"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type KeycloakClient struct {
	openidClient *openid.OpenidClient

	server string
	realm  string
}

func (c *KeycloakClient) getHttpClient(
	ctx context.Context, useCtxBearer bool,
) (*http.Client, error) {
	if useCtxBearer {
		return seedbearer.InterceptBearerTransport(http.DefaultClient), nil
	}
	client, err := c.openidClient.Client(ctx, nil)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return client, nil
}

// GetRealm returns the Keycloak admin REST representation of the client's realm.
// The request authenticates with the client's configured transport, which must
// carry a token holding the realm-management view-realm role.
func (c *KeycloakClient) GetRealm(
	ctx context.Context, useCtxBearer bool,
) (*RealmRepresentation, error) {
	httpClient, err := c.getHttpClient(ctx, useCtxBearer)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	requestUrl := c.server + "/admin/realms/" + url.PathEscape(c.realm)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUrl, nil)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	defer response.Body.Close()
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, seederr.WrapErrorf("failed to get keycloak realm: status %d, body: %s", response.StatusCode, string(responseBytes))
	}

	realm := &RealmRepresentation{}
	err = json.Unmarshal(responseBytes, realm)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return realm, nil
}

// GetUser fetches a single user by Keycloak uuid. The request
// authenticates with the client's configured transport, which must carry a
// token holding the realm-management view-users role. A missing user surfaces
// as fs.ErrNotExist.
func (c *KeycloakClient) GetUser(
	ctx context.Context, useCtxBearer bool, userUuid string,
) (*UserRepresentation, error) {
	httpClient, err := c.getHttpClient(ctx, useCtxBearer)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	requestUrl := c.server + "/admin/realms/" + url.PathEscape(c.realm) +
		"/users/" + url.PathEscape(userUuid)
	seedlog.Debugf("Request: %s", requestUrl)

	request, err := http.NewRequestWithContext(
		ctx, http.MethodGet, requestUrl, nil)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	defer response.Body.Close()
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if response.StatusCode == http.StatusNotFound {
		return nil, seederr.Wrap(fs.ErrNotExist)
	}
	if response.StatusCode != http.StatusOK {
		return nil, seederr.WrapErrorf("failed to get keycloak user: status %d, body: %s", response.StatusCode, string(responseBytes))
	}
	seedlog.Debugf("Response: %s", string(responseBytes))

	user := &UserRepresentation{}
	err = json.Unmarshal(responseBytes, user)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return user, nil
}

// GetKeycloakUser resolves the Keycloak username to the user's
// Keycloak uuid. The request authenticates as the login user via the
// bearer transport, so that user must hold the realm-management view-users role.
func (c *KeycloakClient) GetUserByHandle(
	ctx context.Context, useCtxBearer bool, userHandle string,
) (*UserRepresentation, error) {
	httpClient, err := c.getHttpClient(ctx, useCtxBearer)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	requestUrl := c.server + "/admin/realms/" + url.PathEscape(c.realm)
	query := url.Values{}
	query.Set("username", userHandle)
	query.Set("exact", "true")
	requestUrl += "/users?" + query.Encode()
	seedlog.Debugf("Request: %s", requestUrl)

	request, err := http.NewRequestWithContext(
		ctx, http.MethodGet, requestUrl, nil)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	defer response.Body.Close()
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, seederr.WrapErrorf("failed to lookup keycloak user: status %d, body: %s", response.StatusCode, string(responseBytes))
	}
	seedlog.Debugf("Response: %s", string(responseBytes))

	users := []UserRepresentation{}
	err = json.Unmarshal(responseBytes, &users)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if len(users) == 0 {
		return nil, seederr.Wrap(fs.ErrNotExist)
	}
	if len(users) > 1 {
		return nil, seederr.WrapErrorf("multiple users found for handle: %s", userHandle)
	}
	return &users[0], nil
}

// usersPageSize bounds each users page. Keycloak caps the page size server-side,
// so ListUsers pages through the realm until a short page signals the end.
const usersPageSize = 100

// ListUsers returns every enabled user in the realm, skipping client service
// accounts. It walks through all the pages, so the result is complete no matter
// how many users there are. The request authenticates with the client's
// configured transport, which must carry a token that holds the
// realm-management view-users role.
func (c *KeycloakClient) ListUsers(
	ctx context.Context, useCtxBearer bool,
) ([]UserRepresentation, error) {
	httpClient, err := c.getHttpClient(ctx, useCtxBearer)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	users := []UserRepresentation{}
	for first := 0; ; first += usersPageSize {
		requestUrl := c.server + "/admin/realms/" + url.PathEscape(c.realm)
		query := url.Values{}
		query.Set("enabled", "true")
		query.Set("first", strconv.Itoa(first))
		query.Set("max", strconv.Itoa(usersPageSize))
		// The search parameter is required to exclude service accounts
		// See: https://github.com/keycloak/keycloak/blob/26.6.4/services/src/main/java/org/keycloak/services/resources/admin/UsersResource.java#L307
		query.Set("search", "*")
		requestUrl += "/users?" + query.Encode()
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUrl, nil)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		response, err := httpClient.Do(request)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		responseBytes, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		if response.StatusCode != http.StatusOK {
			return nil, seederr.WrapErrorf("failed to list keycloak users: status %d, body: %s", response.StatusCode, string(responseBytes))
		}

		page := []UserRepresentation{}
		err = json.Unmarshal(responseBytes, &page)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		users = append(users, page...)
		if len(page) < usersPageSize {
			return users, nil
		}
	}
}

// GenerateUserPassword sets a fresh random password on the user and
// returns it. The password is non-temporary so it can immediately back a
// password grant, and is expected to be removed afterwards.
func (c *KeycloakClient) GenerateUserPassword(
	ctx context.Context, useCtxBearer bool, userUuid string,
) (string, error) {
	httpClient, err := c.getHttpClient(ctx, useCtxBearer)
	if err != nil {
		return "", seederr.Wrap(err)
	}

	randomBytes := make([]byte, 16)
	_, err = rand.Read(randomBytes)
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

	requestUrl := c.server + "/admin/realms/" + url.PathEscape(c.realm) +
		"/users/" + url.PathEscape(userUuid) + "/reset-password"
	request, err := http.NewRequestWithContext(
		ctx, http.MethodPut, requestUrl, bytes.NewReader(credentialBytes))
	if err != nil {
		return "", seederr.Wrap(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
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
func (c *KeycloakClient) RemoveUserPassword(
	ctx context.Context, useCtxBearer bool, userUuid string,
) error {
	httpClient, err := c.getHttpClient(ctx, useCtxBearer)
	if err != nil {
		return seederr.Wrap(err)
	}

	requestUrl := c.server + "/admin/realms/" + url.PathEscape(c.realm) +
		"/users/" + url.PathEscape(userUuid) + "/credentials"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUrl, nil)
	if err != nil {
		return seederr.Wrap(err)
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer response.Body.Close()
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return seederr.Wrap(err)
	}
	if response.StatusCode != http.StatusOK {
		return seederr.WrapErrorf("failed to list keycloak user credentials: status %d, body: %s", response.StatusCode, string(responseBytes))
	}

	credentials := []CredentialRepresentation{}
	err = json.Unmarshal(responseBytes, &credentials)
	if err != nil {
		return seederr.Wrap(err)
	}

	for _, credential := range credentials {
		if credential.Type != "password" {
			continue
		}
		deleteUrl := c.server + "/admin/realms/" + url.PathEscape(c.realm) +
			"/users/" + url.PathEscape(userUuid) + "/credentials/" + url.PathEscape(credential.Id)
		deleteRequest, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteUrl, nil)
		if err != nil {
			return seederr.Wrap(err)
		}
		deleteResponse, err := httpClient.Do(deleteRequest)
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

func CreateKeycloakClient(openidClient *openid.OpenidClient) (*KeycloakClient, error) {
	parsedUrl, err := url.Parse(openidClient.DiscoveryUrl())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	server := parsedUrl.Scheme + "://" + parsedUrl.Host
	match := discoveryPathRegex.FindStringSubmatch(parsedUrl.Path)
	if match == nil {
		return nil, seederr.WrapErrorf("invalid keycloak discovery URL: %v", parsedUrl)
	}
	realm := strings.TrimSpace(match[1])
	return &KeycloakClient{
		openidClient: openidClient,

		server: server,
		realm:  realm,
	}, nil
}
