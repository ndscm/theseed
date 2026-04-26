package seedgrpc

import (
	"net"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/soheilhy/cmux"
)

var flagEnableHttp = seedflag.DefineBool("enable_http", true, "")
var flagEnableHttps = seedflag.DefineBool("enable_https", false, "")
var flagHttpsCertificateKeyFile = seedflag.DefineString("https_certificate_key_file", "", "")
var flagHttpsCertificateFile = seedflag.DefineString("https_certificate_file", "", "")

func WithCommonInterceptors(options ...connect.Interceptor) connect.Option {
	interceptors := []connect.Interceptor{}
	interceptors = append(interceptors, options...)
	interceptors = append(interceptors, NewLogInterceptor())
	interceptors = append(interceptors, validate.NewInterceptor())
	return connect.WithInterceptors(interceptors...)
}

func goServeHttp(listener net.Listener, handler http.Handler) {
	protocolVersions := &http.Protocols{}
	protocolVersions.SetHTTP1(true)
	protocolVersions.SetUnencryptedHTTP2(true)
	server := &http.Server{
		Handler:   handler,
		Protocols: protocolVersions,
	}
	seedlog.Infof("Serving http at %v", listener.Addr())
	err := server.Serve(listener)
	if err != nil {
		seedlog.Errorf("[http] %v", err)
	}
}

func goServeHttps(listener net.Listener, handler http.Handler) {
	protocolVersions := &http.Protocols{}
	protocolVersions.SetHTTP1(true)
	protocolVersions.SetHTTP2(true)
	server := &http.Server{
		Handler:   handler,
		Protocols: protocolVersions,
	}
	seedlog.Infof("Serving https at %v", listener.Addr())
	err := server.ServeTLS(listener,
		flagHttpsCertificateFile.Get(), flagHttpsCertificateKeyFile.Get())
	if err != nil {
		seedlog.Errorf("[https] %v", err)
	}
}

func cmuxListenAndServe(addr string, handler http.Handler) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return seederr.Wrap(err)
	}
	m := cmux.New(listener)

	// Sniff http requests
	httpListener := m.Match(cmux.HTTP1Fast())
	anyListener := m.Match(cmux.Any())

	if flagEnableHttp.Get() {
		go goServeHttp(httpListener, handler)
	}
	if flagEnableHttps.Get() {
		go goServeHttps(anyListener, handler)
	}
	seedlog.Infof("Serving at %v", addr)
	err = m.Serve()
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func ListenAndServe(addr string, handler http.Handler) error {
	if flagEnableHttp.Get() && flagEnableHttps.Get() {
		return cmuxListenAndServe(addr, handler)
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return seederr.Wrap(err)
	}
	if flagEnableHttp.Get() {
		goServeHttp(listener, handler)
		return nil
	}
	if flagEnableHttps.Get() {
		goServeHttps(listener, handler)
		return nil
	}
	return seederr.WrapErrorf("no valid protocol enabled")
}
