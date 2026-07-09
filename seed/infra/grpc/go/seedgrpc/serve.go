package seedgrpc

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/soheilhy/cmux"
)

var flagEnableHttp = seedflag.DefineBool("enable_http", true, "")
var flagEnableHttps = seedflag.DefineBool("enable_https", false, "")
var flagHttpsCertificateKeyFile = seedflag.DefineString("https_certificate_key_file", "", "")
var flagHttpsCertificateFile = seedflag.DefineString("https_certificate_file", "", "")

type SniffServer struct {
	addr    string
	handler http.Handler

	rootListenerMutex sync.Mutex
	rootListener      net.Listener

	httpServer  *http.Server
	httpsServer *http.Server
}

func (s *SniffServer) setRootListener(listener net.Listener) {
	s.rootListenerMutex.Lock()
	defer s.rootListenerMutex.Unlock()
	s.rootListener = listener
}

func (s *SniffServer) closeRootListener() error {
	s.rootListenerMutex.Lock()
	defer s.rootListenerMutex.Unlock()
	if s.rootListener == nil {
		return nil
	}
	return s.rootListener.Close()
}

func (s *SniffServer) serveHttp(listener net.Listener) error {
	seedlog.Infof("Serving http at %v", listener.Addr())
	err := s.httpServer.Serve(listener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return seederr.Wrap(err)
	}
	return nil
}

func (s *SniffServer) serveHttps(listener net.Listener) error {
	seedlog.Infof("Serving https at %v", listener.Addr())
	err := s.httpsServer.ServeTLS(listener,
		flagHttpsCertificateFile.Get(), flagHttpsCertificateKeyFile.Get())
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return seederr.Wrap(err)
	}
	return nil
}

func (s *SniffServer) cmuxServe(listener net.Listener) error {
	m := cmux.New(listener)

	// Sniff http requests
	httpListener := m.Match(cmux.HTTP1Fast())
	anyListener := m.Match(cmux.Any())

	if s.httpServer != nil {
		go func() {
			err := s.serveHttp(httpListener)
			if err != nil {
				seedlog.Errorf("[http] %v", err)
			}
		}()
	}

	if s.httpsServer != nil {
		go func() {
			err := s.serveHttps(anyListener)
			if err != nil {
				seedlog.Errorf("[https] %v", err)
			}
		}()
	}

	seedlog.Infof("Serving cmux at %v", listener.Addr())
	err := m.Serve()
	if err != nil && !errors.Is(err, net.ErrClosed) {
		return seederr.Wrap(err)
	}
	return nil
}

func (s *SniffServer) Serve() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return seederr.Wrap(err)
	}
	s.setRootListener(listener)
	if s.httpServer != nil && s.httpsServer != nil {
		err = s.cmuxServe(listener)
		if err != nil {
			return seederr.Wrap(err)
		}
	} else if s.httpServer != nil {
		err = s.serveHttp(listener)
		if err != nil {
			return seederr.Wrap(err)
		}
	} else if s.httpsServer != nil {
		err = s.serveHttps(listener)
		if err != nil {
			return seederr.Wrap(err)
		}
	} else {
		return seederr.WrapErrorf("no valid protocol enabled")
	}
	return nil
}

func (s *SniffServer) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	if s.httpsServer != nil {
		err := s.httpsServer.Shutdown(ctx)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	err := s.closeRootListener()
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func (s *SniffServer) Close() error {
	if s.httpServer != nil {
		err := s.httpServer.Close()
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	if s.httpsServer != nil {
		err := s.httpsServer.Close()
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	err := s.closeRootListener()
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func CreateServer(addr string, handler http.Handler) (*SniffServer, error) {
	s := &SniffServer{
		addr:    addr,
		handler: handler,
	}
	if flagEnableHttp.Get() {
		protocolVersions := &http.Protocols{}
		protocolVersions.SetHTTP1(true)
		protocolVersions.SetUnencryptedHTTP2(true)
		s.httpServer = &http.Server{
			Handler:   s.handler,
			Protocols: protocolVersions,
		}
	}
	if flagEnableHttps.Get() {
		protocolVersions := &http.Protocols{}
		protocolVersions.SetHTTP1(true)
		protocolVersions.SetHTTP2(true)
		s.httpsServer = &http.Server{
			Handler:   s.handler,
			Protocols: protocolVersions,
		}
	}
	if !flagEnableHttp.Get() && !flagEnableHttps.Get() {
		return nil, seederr.WrapErrorf("no valid protocol enabled")
	}
	return s, nil
}

func ListenAndServe(addr string, handler http.Handler) error {
	s, err := CreateServer(addr, handler)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = s.Serve()
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
