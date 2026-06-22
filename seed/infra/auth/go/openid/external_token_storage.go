package openid

import (
	"context"
	"sync"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/oauth2"
)

type ExternalTokenStorage interface {
	Get(ctx context.Context, key string) (string, error)
	Update(ctx context.Context, change map[string]string) error
}

type ExternalTokenStorageTokenSource struct {
	mutex   sync.Mutex
	ctx     context.Context
	prefix  string
	next    oauth2.TokenSource
	storage ExternalTokenStorage
	last    *oauth2.Token
}

func (s *ExternalTokenStorageTokenSource) Token() (*oauth2.Token, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	newToken, err := s.next.Token()
	if err != nil {
		return nil, err
	}
	if s.last == nil || s.last.AccessToken != newToken.AccessToken {
		if s.storage == nil {
			return nil, seederr.WrapErrorf("storage is nil")
		}
		err := s.storage.Update(s.ctx, map[string]string{
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

var _ oauth2.TokenSource = (*ExternalTokenStorageTokenSource)(nil)

func createExternalTokenStorageTokenSource(
	ctx context.Context,
	prefix string,
	oauth2Config *oauth2.Config,
	storage ExternalTokenStorage,
	initial *oauth2.Token,
) (*ExternalTokenStorageTokenSource, error) {
	if oauth2Config == nil {
		return nil, seederr.WrapErrorf("oauth2 config is nil")
	}
	if storage == nil {
		return nil, seederr.WrapErrorf("storage is nil")
	}
	if initial == nil {
		accessToken, err := storage.Get(ctx, prefix+"access_token")
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		refreshToken, err := storage.Get(ctx, prefix+"refresh_token")
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		expiryString, err := storage.Get(ctx, prefix+"expiry")
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		// Default to an already-expired time so a missing or corrupt stored
		// expiry forces a refresh. A zero time.Time would make oauth2 treat the
		// token as non-expiring and never refresh it.
		expiry := time.Now().Add(-time.Minute)
		if expiryString != "" {
			parsedExpiry, err := time.Parse(time.RFC3339Nano, expiryString)
			if err != nil {
				seedlog.Warnf("Invalid token expiry. expiry=%v", expiryString)
			} else {
				expiry = parsedExpiry
			}
		}
		initial = &oauth2.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			Expiry:       expiry,
		}
	} else {
		err := storage.Update(ctx, map[string]string{
			prefix + "access_token":  initial.AccessToken,
			prefix + "refresh_token": initial.RefreshToken,
			prefix + "expiry":        initial.Expiry.Format(time.RFC3339Nano),
		})
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	}
	tokenSource := &ExternalTokenStorageTokenSource{
		ctx:     ctx,
		prefix:  prefix,
		next:    oauth2Config.TokenSource(ctx, initial),
		storage: storage,
		last:    initial,
	}
	return tokenSource, nil
}
