package seedgrpc

import (
	"net/http"

	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type GrpcMux struct {
	http.ServeMux
}

func (cls *GrpcMux) Register(path string, handler http.Handler) error {
	cls.Handle(path, handler)
	seedlog.Infof("Service registered: %v", path)
	return nil
}
