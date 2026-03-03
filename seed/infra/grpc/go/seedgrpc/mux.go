package seedgrpc

import (
	"net/http"
	"strings"

	"connectrpc.com/grpchealth"
	"github.com/ndscm/theseed/seed/infra/http/go/seedjwt"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type GrpcMux struct {
	http.ServeMux
	serviceNames  []string
	healthChecker *grpchealth.StaticChecker
}

func (cls *GrpcMux) Register(path string, handler http.Handler) error {
	handler = seedjwt.InterceptJwtMiddleware(handler)
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
