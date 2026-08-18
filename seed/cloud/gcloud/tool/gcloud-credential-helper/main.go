// Command gcloud-credential-helper prints a Google Cloud access token obtained
// from Application Default Credentials (ADC) in one of several formats (raw, the
// serialized OAuth2 token, or the Bazel --credential_helper wire format). It is
// primarily a Bazel credential helper for the GCS remote cache, invoked as
// `gcloud-credential-helper get` with a request on stdin and the reply on stdout.
//
// It replaces shelling out to `gcloud auth application-default
// print-access-token`: the Google auth library resolves ADC (the gcloud
// application-default file, a service account key, or the metadata server) and
// mints — and refreshes — the access token itself.
package main

import (
	"context"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var flagScopes = seedflag.DefineStringList(
	"scopes", []string{"https://www.googleapis.com/auth/cloud-platform"},
	"OAuth scopes to request for the ADC token",
)
var flagTokenCacheFile = seedflag.DefineString(
	"token_cache_file", "~/.seed/login/gcloud_application_default_token.json",
	`Token cache file; a leading ~/ is expanded to the home directory. Defaults `+
		`to ~/.seed/login/gcloud_application_default_token.json. Set to empty to disable `+
		`caching, so a fresh token is minted every call and nothing is written to `+
		`disk. A caller may pre-write a token here (a serialized oauth2.Token `+
		`JSON) to feed a cached token, and may point at an unlinked fd to keep the `+
		`token off disk. Change this when --scopes changes, since the cache is not `+
		`keyed by scope`,
)
var flagOutputFormat = seedflag.DefineString(
	"output_format", "access_token",
	`Output format written to stdout (no trailing newline): `+
		`"access_token" (the raw bearer token), `+
		`"oauth2" (the serialized OAuth2 token: access token, expiry, ...), `+
		`or "bazel" (the JSON a Bazel --credential_helper consumes)`,
)

// BazelOutput is the Bazel credential helper reply written to stdout. The
// headers are attached to the outgoing request; expires lets Bazel cache the
// token until it lapses instead of invoking the helper on every call.
type BazelOutput struct {
	Headers map[string][]string `json:"headers"`
	Expires string              `json:"expires,omitempty"`
}

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

// obtainToken returns a valid ADC access token, optionally cached on disk per
// token_cache_file. A cached token that stays valid past the expiry margin is
// reused; otherwise a fresh token is resolved from Application Default
// Credentials and, when caching is enabled, persisted.
//
// The Google auth library only reuses a token in memory, which cannot help a
// per-call helper process, so the on-disk cache is what keeps repeated
// invocations from refreshing against the token endpoint every time.
func obtainToken(ctx context.Context) (*oauth2.Token, error) {
	// token_cache_file is expanded (a leading ~/ becomes the home directory) into
	// the cache path. Empty disables caching, so nothing is read from or written
	// to disk.
	tokenPath := ""
	tokenCacheFile := flagTokenCacheFile.Get()
	if tokenCacheFile != "" {
		tokenPath = tokenCacheFile
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

	if !fresh {
		tokenSource, err := google.DefaultTokenSource(ctx, flagScopes.Get()...)
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
	default:
		return nil, seederr.WrapErrorf(
			"unknown output_format: %q", format)
	}
	return output, nil
}

func run() error {
	_, err := seedinit.Initialize(
		seedinit.WithEnvPrefix("GCLOUD_CREDENTIAL_HELPER_"),
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
