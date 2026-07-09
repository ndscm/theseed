package seedgrpc

import (
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/validate"
	"github.com/ndscm/theseed/seed/infra/grpc/go/grpclog"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

func WithCommonInterceptors(options ...connect.Interceptor) connect.Option {
	interceptors := []connect.Interceptor{}
	interceptors = append(interceptors, options...)
	interceptors = append(interceptors, grpclog.NewLogInterceptor())
	interceptors = append(interceptors, validate.NewInterceptor())
	return connect.WithInterceptors(interceptors...)
}

type GrpcMux struct {
	http.ServeMux

	interceptors []func(http.Handler) http.Handler

	serviceNames  []string
	healthChecker *grpchealth.StaticChecker
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

func CreateGrpcMux(interceptors ...func(http.Handler) http.Handler) (*GrpcMux, error) {
	m := &GrpcMux{}
	m.interceptors = interceptors
	return m, nil
}
