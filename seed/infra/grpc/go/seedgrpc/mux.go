package seedgrpc

import (
	"net/http"
	"strings"

	"connectrpc.com/grpchealth"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type GrpcMux struct {
	http.ServeMux

	interceptors []func(http.Handler) http.Handler

	serviceNames  []string
	healthChecker *grpchealth.StaticChecker
}

func CreateGrpcMux(interceptors ...func(http.Handler) http.Handler) (*GrpcMux, error) {
	m := &GrpcMux{}
	m.interceptors = interceptors
	return m, nil
}

func (cls *GrpcMux) Register(path string, handler http.Handler) error {
	for _, interceptor := range cls.interceptors {
		handler = interceptor(handler)
	}
	cls.Handle(path, handler)
	seedlog.Infof("Service registered: %v", path)
	cls.serviceNames = append(cls.serviceNames, strings.Trim(path, "/"))
	return nil
}

func (cls *GrpcMux) Ready() (http.Handler, error) {
	cls.healthChecker = grpchealth.NewStaticChecker(
		cls.serviceNames...,
	)
	cls.Handle(grpchealth.NewHandler(cls.healthChecker))
	return cls, nil
}
