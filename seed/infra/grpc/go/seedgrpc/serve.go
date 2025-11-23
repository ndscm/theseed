package seedgrpc

import (
	"net/http"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

func ListenAndServe(addr string, handler http.Handler) error {
	protocolVersions := &http.Protocols{}
	protocolVersions.SetHTTP1(true)
	protocolVersions.SetUnencryptedHTTP2(true)
	server := &http.Server{
		Addr:      addr,
		Handler:   handler,
		Protocols: protocolVersions,
	}
	seedlog.Infof("Serving at %v", addr)
	err := server.ListenAndServe()
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
