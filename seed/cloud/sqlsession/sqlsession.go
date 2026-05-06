package sqlsession

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ndscm/theseed/seed/cloud/sqlsession/database/ent"
	"github.com/ndscm/theseed/seed/cloud/sqlsession/database/ent/session"
	"github.com/ndscm/theseed/seed/cloud/sqlsession/database/sqlsessiondb"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/http/go/seedsession"
)

// sessionDuration is 24 hours + 6 hour buffer so users keep their session
// alive with any daily visit without racing against an exact 24h deadline.
const sessionDuration = 30 * time.Hour

type sqlSessionAdapter struct {
	client    *ent.Client
	sessionId string
	data      map[string]string
}

func (s *sqlSessionAdapter) SessionId() string {
	return s.sessionId
}

func (s *sqlSessionAdapter) Init(ctx context.Context, sessionId string, responseHeaders http.Header) error {
	if sessionId != "" {
		sessionUuid, err := uuid.Parse(sessionId)
		if err != nil {
			sessionId = ""
		} else {
			row, err := s.client.Session.Get(ctx, sessionUuid)
			if err != nil {
				if ent.IsNotFound(err) {
					sessionId = ""
				} else {
					return seederr.Wrap(err)
				}
			} else if time.Now().After(row.ExpireTime) {
				err := s.client.Session.DeleteOneID(sessionUuid).Exec(ctx)
				if err != nil {
					return seederr.Wrap(err)
				}
				sessionId = ""
			}
		}
	}
	if sessionId == "" {
		dataBytes, err := json.Marshal(map[string]string{})
		if err != nil {
			return seederr.Wrap(err)
		}
		raw := json.RawMessage(dataBytes)
		expires := time.Now().Add(sessionDuration)
		row, err := s.client.Session.Create().SetData(&raw).SetExpireTime(expires).Save(ctx)
		if err != nil {
			return seederr.Wrap(err)
		}
		sessionId = row.ID.String()
		responseHeaders.Add("Set-Cookie", seedsession.WrapCookieString(sessionId, expires, true))
	}
	s.sessionId = sessionId
	return nil
}

func (s *sqlSessionAdapter) Refresh(ctx context.Context, responseHeaders http.Header) error {
	if s.sessionId == "" {
		return seederr.WrapErrorf("session not initialized")
	}
	sessionUuid, err := uuid.Parse(s.sessionId)
	if err != nil {
		return seederr.Wrap(err)
	}
	expires := time.Now().Add(sessionDuration)
	err = s.client.Session.UpdateOneID(sessionUuid).SetExpireTime(expires).Exec(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}
	responseHeaders.Add("Set-Cookie", seedsession.WrapCookieString(s.sessionId, expires, true))
	return nil
}

func (s *sqlSessionAdapter) Reload(ctx context.Context) error {
	sessionUuid, err := uuid.Parse(s.sessionId)
	if err != nil {
		return seederr.Wrap(err)
	}
	row, err := s.client.Session.Get(ctx, sessionUuid)
	if err != nil {
		return seederr.Wrap(err)
	}
	data := map[string]string{}
	if row.Data != nil {
		err = json.Unmarshal(*row.Data, &data)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	s.data = data
	return nil
}

func (s *sqlSessionAdapter) Get(ctx context.Context, key string) (string, error) {
	if s.data == nil {
		err := s.Reload(ctx)
		if err != nil {
			return "", seederr.Wrap(err)
		}
	}
	if value, ok := s.data[key]; ok {
		return value, nil
	}
	return "", nil
}

func (s *sqlSessionAdapter) Update(ctx context.Context, change map[string]string) error {
	sessionUuid, err := uuid.Parse(s.sessionId)
	if err != nil {
		return seederr.Wrap(err)
	}
	tx, err := s.client.BeginTx(ctx, nil)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer tx.Rollback()
	row, err := tx.Session.Query().Where(session.ID(sessionUuid)).ForUpdate().Only(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}
	data := map[string]string{}
	if row.Data != nil {
		err = json.Unmarshal(*row.Data, &data)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	maps.Copy(data, change)
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return seederr.Wrap(err)
	}
	raw := json.RawMessage(dataBytes)
	err = tx.Session.UpdateOneID(sessionUuid).SetData(&raw).Exec(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}
	return tx.Commit()
}

func CreateSqlSessionInitializer() (seedsession.SessionInitializer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := sqlsessiondb.Open(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	err = client.Schema.Create(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	sessionInitializer := func() seedsession.SessionAdapter {
		return &sqlSessionAdapter{client: client}
	}
	return sessionInitializer, nil
}
