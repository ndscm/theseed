package seedgzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

type gzipMiddleware struct {
	next http.Handler
}

func (g *gzipMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		g.next.ServeHTTP(w, r)
		return
	}
	w.Header().Set("Content-Encoding", "gzip")
	gzipWriter := gzip.NewWriter(w)
	defer gzipWriter.Close()
	g.next.ServeHTTP(&gzipResponseWriter{Writer: gzipWriter, ResponseWriter: w}, r)
}

func InterceptGzipMiddleware(next http.Handler) http.Handler {
	return &gzipMiddleware{next: next}
}
