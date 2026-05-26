package escalate

import (
	"net/http"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	redirectUrl := *r.URL
	redirectUrl.Scheme = "https"
	redirectUrl.Host = r.Host
	http.Redirect(w, r, redirectUrl.String(), http.StatusTemporaryRedirect)
}

var _ http.Handler = http.HandlerFunc(ServeHTTP)

type EscalateHandler struct{}

func (h *EscalateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ServeHTTP(w, r)
}

var _ http.Handler = (*EscalateHandler)(nil)

func NewEscalateHandler() *EscalateHandler {
	return &EscalateHandler{}
}
