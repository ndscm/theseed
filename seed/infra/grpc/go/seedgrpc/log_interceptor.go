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
		seedlog.Infof("Grpc unary: %s", request.Spec().Procedure)
		seedlog.Debugf("Grpc headers: %+v", request.Header())
		seedlog.Debugf("Grpc request: (%T) %+v", request.Any(), request.Any())
		seedErrorCode := uint32(0)
		response, err := next(ctx, request)
		if err != nil {
			seedlog.Errorf("%s error: %v", request.Spec().Procedure, err)
			seedErr := &seederr.SeedError{}
			if errors.As(err, &seedErr) {
				seedErrorCode = seedErr.Code()
				err = seedErr.Unwrap()
			}
		}
		if !reflect.ValueOf(response).IsNil() {
			seedlog.Debugf("Grpc response: (%T) %+v", response.Any(), response.Any())
		}
		grpcErrorCode := seedErrorCode & 0xff
		if grpcErrorCode != 0 {
			return response, connect.NewError(connect.Code(grpcErrorCode), err)
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
