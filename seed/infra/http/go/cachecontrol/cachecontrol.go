package cachecontrol

import (
	"fmt"
	"net/http"
	"time"
)

type cacheControlMiddleware struct {
	next    http.Handler
	seconds int
}

func (cls *cacheControlMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%v", cls.seconds))
	expires := time.Now().Add(time.Duration(cls.seconds) * time.Second)
	w.Header().Set("Expires", expires.Format(http.TimeFormat))
	cls.next.ServeHTTP(w, r)
}

func InterceptCacheControlMiddleware(next http.Handler, seconds int) http.Handler {
	return &cacheControlMiddleware{next: next, seconds: seconds}
}
