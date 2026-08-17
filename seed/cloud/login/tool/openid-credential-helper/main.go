// Command openid-credential-helper produces an OpenID Connect access token and
// prints it to stdout in one of several formats (raw, oauth2 JSON, or the Bazel
// --credential_helper wire format). It is primarily a Bazel credential helper,
// invoked as `openid-credential-helper get` with a request on stdin and the
// reply on stdout.
//
// The token comes from the client-credentials grant and, by default, is cached
// under ~/.seed/login (the layout devicelogin writes) via --token_cache_file; set
// that flag empty to disable caching and keep the token off disk. A caller that
// wants to feed a pre-obtained token writes it to the cache file itself, which
// may be an unlinked fd so the token is never visible on disk.
package main

import (
	"context"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/oauth2"
)

var flagOpenidClientId = seedflag.DefineString(
	"openid_client_id", "",
	"Client ID for OpenID Connect",
)
var flagOpenidClientSecret = seedflag.DefineSecret(
	"openid_client_secret",
	"Client Secret for OpenID Connect",
)
var flagOpenidScopes = seedflag.DefineStringList(
	"openid_scopes", []string{"openid", "profile", "email"},
	"OAuth scopes to request; empty requests the client's default scopes",
)
var flagTokenCacheFile = seedflag.DefineString(
	"token_cache_file", "~/.seed/login/<openid_client_id>.json",
	`Token cache file; a leading ~/ is expanded to the home directory and `+
		`<openid_client_id> to the OpenID client id. Defaults to `+
		`~/.seed/login/<openid_client_id>.json, alongside devicelogin's per-service `+
		`files. Set to empty to disable caching, so a fresh token is minted every `+
		`call and nothing is written to disk. A caller may pre-write a token here `+
		`(a serialized oauth2.Token JSON) to feed a cached token, and may point at `+
		`an unlinked fd to keep the token off disk`,
)
var flagOutputFormat = seedflag.DefineString(
	"output_format", "access_token",
	`Output format written to stdout (no trailing newline): `+
		`"access_token" (the raw bearer token), `+
		`"oauth2" (the serialized OAuth2 token: access token, refresh token, expiry, ...), `+
		`"bazel" (the JSON a Bazel --credential_helper consumes), `+
		`or "gcloud" (the JSON an executable-sourced gcloud external account `+
		`credential consumes, carrying the access token as an OIDC JWT)`,
)

// expiryMargin keeps a cached token in reserve before it actually expires, so a
// token handed back to Bazel stays valid for the request it is used on.
const expiryMargin = time.Minute

// loadToken reads the cached token, returning a nil token when caching is
// disabled (empty path) or no cache file exists yet.
func loadToken(path string) (*oauth2.Token, error) {
	if path == "" {
		return nil, nil
	}
	tokenBytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, seederr.Wrap(err)
	}
	token := &oauth2.Token{}
	err = json.Unmarshal(tokenBytes, token)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return token, nil
}

// saveToken writes the token to path as a serialized oauth2.Token, creating
// parent directories. It is a no-op when caching is disabled (empty path).
func saveToken(path string, token *oauth2.Token) error {
	if path == "" {
		return nil
	}
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.WriteFile(path, tokenBytes, 0600)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// newOpenidClient constructs the OpenID client from the client flags, requiring
// the client id and discovery URL.
func newOpenidClient() (*openid.OpenidClient, error) {
	clientId := flagOpenidClientId.Get()
	if clientId == "" {
		return nil, seederr.WrapErrorf("openid_client_id is required")
	}
	clientSecret, err := flagOpenidClientSecret.LoadString()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	discoveryUrl := openid.OpenidDiscoveryUrlFlag()
	if discoveryUrl == "" {
		return nil, seederr.WrapErrorf("openid_discovery_url is required")
	}
	return openid.NewOpenidClient(discoveryUrl, clientId, clientSecret), nil
}

// obtainToken returns a valid access token from the client-credentials grant,
// optionally cached on disk per token_cache_file. A cached token that stays valid
// past the expiry margin is reused; otherwise a fresh token is minted and, when
// caching is enabled, persisted.
func obtainToken(ctx context.Context) (*oauth2.Token, error) {
	openidClient, err := newOpenidClient()
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	// token_cache_file defaults to the per-client file kept alongside
	// devicelogin's per-service files under ~/.seed/login; <openid_client_id> in it is
	// replaced with the client id and a leading ~/ with the home directory.
	// Empty disables caching, so nothing is read from or written to disk.
	tokenPath := ""
	if flagTokenCacheFile.Get() != "" {
		tokenPath = strings.ReplaceAll(flagTokenCacheFile.Get(), "<openid_client_id>", openidClient.ClientId())
		if strings.HasPrefix(tokenPath, "~/") {
			userHome, err := os.UserHomeDir()
			if err != nil {
				return nil, seederr.Wrap(err)
			}
			tokenPath = filepath.Join(userHome, tokenPath[2:])
		}
	}

	token, err := loadToken(tokenPath)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	// A cached token can be reused when it exists and stays valid past the
	// expiry margin.
	fresh := token != nil &&
		token.AccessToken != "" &&
		token.Expiry.After(time.Now().Add(expiryMargin))

	// Bazel spawns a fresh helper process per call, so the on-disk cache is
	// what keeps this from hitting the token endpoint every time. Mint (and
	// persist, when caching is enabled) a new token only when the cached one
	// is missing or about to expire.
	if !fresh {
		tokenSource, err := openidClient.GetTokenSource(ctx, flagOpenidScopes.Get())
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		token, err = tokenSource.Token()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		err = saveToken(tokenPath, token)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	}
	return token, nil
}

// BazelOutput is the Bazel credential helper reply written to stdout. The
// headers are attached to the outgoing request; expires lets Bazel cache the
// token until it lapses instead of invoking the helper on every call.
type BazelOutput struct {
	Headers map[string][]string `json:"headers"`
	Expires string              `json:"expires,omitempty"`
}

// GcloudOutput is the JSON an executable-sourced gcloud external account
// credential expects on stdout. It carries the access token as an OIDC JWT that
// gcloud exchanges for Google credentials; expiration_time (Unix seconds) lets
// gcloud cache it.
// See https://cloud.google.com/iam/docs/workload-download-cred-and-grant-access.
type GcloudOutput struct {
	Version        int    `json:"version"`
	Success        bool   `json:"success"`
	TokenType      string `json:"token_type"`
	IdToken        string `json:"id_token"`
	ExpirationTime int64  `json:"expiration_time,omitempty"`
}

// formatToken renders the token for stdout in the requested output format,
// without a trailing newline in any case.
func formatToken(format string, token *oauth2.Token) ([]byte, error) {
	output := []byte(nil)
	switch format {
	case "access_token":
		output = []byte(token.AccessToken)
	case "oauth2":
		marshaled, err := json.Marshal(token)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		output = marshaled
	case "bazel":
		response := &BazelOutput{
			Headers: map[string][]string{
				"Authorization": {"Bearer " + token.AccessToken},
			},
		}
		if !token.Expiry.IsZero() {
			response.Expires = token.Expiry.UTC().Format(time.RFC3339)
		}
		marshaled, err := json.Marshal(response)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		output = marshaled
	case "gcloud":
		credential := &GcloudOutput{
			Version:   1,
			Success:   true,
			TokenType: "urn:ietf:params:oauth:token-type:jwt",
			IdToken:   token.AccessToken,
		}
		if !token.Expiry.IsZero() {
			credential.ExpirationTime = token.Expiry.Unix()
		}
		marshaled, err := json.Marshal(credential)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		output = marshaled
	default:
		return nil, seederr.WrapErrorf(
			"unknown output_format: %q", format)
	}
	return output, nil
}

func run() error {
	_, err := seedinit.Initialize(
		seedinit.WithEnvPrefix("CREDENTIAL_HELPER_"),
		seedinit.WithFallbackEnvPrefix("SEED_"),
	)
	if err != nil {
		return seederr.Wrap(err)
	}

	ctx := context.Background()

	token, err := obtainToken(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}

	output, err := formatToken(flagOutputFormat.Get(), token)
	if err != nil {
		return seederr.Wrap(err)
	}
	_, err = os.Stdout.Write(output)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		seedlog.Errorf("%v", err)
		os.Exit(1)
	}
}
