package seedsession

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type memorySession struct {
	data    map[string]string
	expires time.Time
}

var memorySessions map[string]memorySession = map[string]memorySession{}

type memorySessionAdapter struct {
	sessionId string
	data      map[string]string
}

func (cls *memorySessionAdapter) SessionId() string {
	return cls.sessionId
}

func (cls *memorySessionAdapter) Init(ctx context.Context, sessionId string, responseHeaders http.Header) error {
	if sessionId != "" {
		session, ok := memorySessions[sessionId]
		if ok {
			if time.Now().After(session.expires) {
				delete(memorySessions, sessionId)
				sessionId = ""
			}
		} else {
			sessionId = ""
		}
	}
	if sessionId == "" {
		b := make([]byte, 32)
		_, err := rand.Read(b)
		if err != nil {
			return seederr.Wrap(err)
		}
		sessionId = fmt.Sprintf("%x", b)
		expires := time.Now().Add(24 * time.Hour)
		memorySessions[sessionId] = memorySession{
			data:    map[string]string{},
			expires: expires.Add(time.Hour),
		}
		responseHeaders.Add("Set-Cookie", WrapCookieString(sessionId, expires, false))
	}
	cls.sessionId = sessionId
	return nil
}

func (cls *memorySessionAdapter) Refresh(ctx context.Context, responseHeaders http.Header) error {
	if cls.sessionId == "" {
		return seederr.WrapErrorf("session not initialized")
	}
	expires := time.Now().Add(24 * time.Hour)
	session, ok := memorySessions[cls.sessionId]
	if !ok {
		return seederr.WrapErrorf("session not found")
	}
	session.expires = expires.Add(time.Hour)
	memorySessions[cls.sessionId] = session
	responseHeaders.Add("Set-Cookie", WrapCookieString(cls.sessionId, expires, false))
	return nil
}

func (cls *memorySessionAdapter) Reload(ctx context.Context) error {
	session, ok := memorySessions[cls.sessionId]
	if !ok {
		return seederr.WrapErrorf("session not found")
	}
	cls.data = session.data
	return nil
}

func (cls *memorySessionAdapter) Get(ctx context.Context, key string) (string, error) {
	if cls.data == nil {
		err := cls.Reload(ctx)
		if err != nil {
			return "", seederr.Wrap(err)
		}
	}
	if value, ok := cls.data[key]; ok {
		return value, nil
	}
	return "", nil
}

func (cls *memorySessionAdapter) Update(ctx context.Context, change map[string]string) error {
	err := cls.Reload(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}
	for key, value := range change {
		cls.data[key] = value
	}
	session, ok := memorySessions[cls.sessionId]
	if !ok {
		return seederr.WrapErrorf("session not found")
	}
	session.data = cls.data
	memorySessions[cls.sessionId] = session
	return nil
}

func MemorySessionInitializer() SessionAdapter {
	return &memorySessionAdapter{}
}
