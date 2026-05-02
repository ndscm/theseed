package loginopenid

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/ndscm/theseed/seed/infra/auth/go/clientopenid"
	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/http/go/seedsession"
	"golang.org/x/oauth2"
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

type UserOpenidProvider struct {
	clientopenid.OpenidProvider
	prefix string
}

func (provider *UserOpenidProvider) Exchange(
	ctx context.Context,
	session seedsession.SessionAdapter,
	origin string,
	code string,
) error {
	oauth2Config, err := provider.GetOauth2Config(ctx, origin)
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

func (provider *UserOpenidProvider) Bearer(
	ctx context.Context,
	session seedsession.SessionAdapter,
) string {
	accessToken, err := session.Get(ctx, provider.prefix+"access_token")
	if err != nil {
		return ""
	}
	// TODO(nagi): check expiry and refresh with refresh token if needed
	return "Bearer " + accessToken
}

func (provider *UserOpenidProvider) Client(
	ctx context.Context,
	session seedsession.SessionAdapter,
	origin string,
) (*http.Client, error) {
	oauth2Config, err := provider.GetOauth2Config(ctx, origin)
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

func (provider *UserOpenidProvider) FetchUserInfo(
	ctx context.Context,
	session seedsession.SessionAdapter,
	origin string,
) (*openid.OpenidUserInfo, error) {
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
	openidUserInfo, err := openid.DecodeOpenidUserInfo(responseBodyBytes)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return openidUserInfo, nil
}

func NewUserOpenidProvider(base *clientopenid.OpenidProvider, prefix string) *UserOpenidProvider {
	return &UserOpenidProvider{
		OpenidProvider: *base,
		prefix:         prefix,
	}
}
