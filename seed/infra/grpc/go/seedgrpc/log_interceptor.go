package seedgrpc

import (
	"context"
	"errors"
	"reflect"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type logInterceptor struct {
}

func (i *logInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		seedlog.Infof("Grpc: %s", request.Spec().Procedure)
		seedlog.Debugf("Headers: %+v", request.Header())
		seedlog.Debugf("Message: (%T) %+v", request.Any(), request.Any())
		response, err := next(ctx, request)
		if err != nil {
			seedlog.Errorf("%s error: %v", request.Spec().Procedure, err)
			seedErr := &seederr.SeedError{}
			if errors.As(err, &seedErr) {
				err = seedErr.Unwrap()
			}
		}
		if !reflect.ValueOf(response).IsNil() {
			seedlog.Debugf("Response: (%T) %+v", response.Any(), response.Any())
		}
		return response, err
	}
}

// WrapStreamingClient implements connect.Interceptor.
func (i *logInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		// TODO(nagi): add logging for streaming RPCs
		return next(ctx, spec)
	}
}

// WrapStreamingHandler implements connect.Interceptor.
func (i *logInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		// TODO(nagi): add logging for streaming RPCs
		return next(ctx, conn)
	}
}

func NewLogInterceptor() *logInterceptor {
	return &logInterceptor{}
}
