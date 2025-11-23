package seedgrpc

import (
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

func WithCommonInterceptors(options ...connect.Interceptor) connect.Option {
	interceptors := []connect.Interceptor{}
	interceptors = append(interceptors, options...)
	interceptors = append(interceptors, NewLogInterceptor())
	interceptors = append(interceptors, validate.NewInterceptor())
	return connect.WithInterceptors(interceptors...)
}

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
