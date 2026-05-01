package seedgrpc

import (
	"context"
	"errors"
	"reflect"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagKeepAuthorizationLog = seedflag.DefineBool("keep_authorization_log", false, "Keep the Authorization header in the debug logs.")

type logInterceptor struct {
}

func (i *logInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		seedlog.Infof("Grpc unary: %s", request.Spec().Procedure)
		logHeaders := request.Header()
		if !flagKeepAuthorizationLog.Get() && logHeaders.Get("Authorization") != "" {
			logHeaders = logHeaders.Clone()
			logHeaders.Set("Authorization", "**REDACTED**")
		}
		seedlog.Debugf("Grpc headers: %+v", logHeaders)
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
		if response != nil {
			v := reflect.ValueOf(response)
			if v.Kind() != reflect.Ptr || !v.IsNil() {
				seedlog.Debugf("Grpc response: (%T) %+v", response.Any(), response.Any())
			}
		}
		grpcErrorCode := seedErrorCode & 0xff
		if grpcErrorCode != 0 {
			return response, connect.NewError(connect.Code(grpcErrorCode), err)
		}
		return response, err
	}
}

type loggingStreamingClientConn struct {
	connect.StreamingClientConn
}

func (c *loggingStreamingClientConn) Send(msg any) error {
	seedlog.Debugf("Grpc client send: (%T) %+v", msg, msg)
	return c.StreamingClientConn.Send(msg)
}

func (c *loggingStreamingClientConn) Receive(msg any) error {
	err := c.StreamingClientConn.Receive(msg)
	if err != nil {
		return err
	}
	seedlog.Debugf("Grpc client receive: (%T) %+v", msg, msg)
	return nil
}

// WrapStreamingClient implements connect.Interceptor.
func (i *logInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		seedlog.Infof("Grpc stream (client): %s", spec.Procedure)
		conn := next(ctx, spec)
		return &loggingStreamingClientConn{StreamingClientConn: conn}
	}
}

type loggingStreamingHandlerConn struct {
	connect.StreamingHandlerConn
}

func (c *loggingStreamingHandlerConn) Receive(msg any) error {
	err := c.StreamingHandlerConn.Receive(msg)
	if err != nil {
		return err
	}
	seedlog.Debugf("Grpc server receive: (%T) %+v", msg, msg)
	return nil
}

func (c *loggingStreamingHandlerConn) Send(msg any) error {
	seedlog.Debugf("Grpc server send: (%T) %+v", msg, msg)
	return c.StreamingHandlerConn.Send(msg)
}

// WrapStreamingHandler implements connect.Interceptor.
func (i *logInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		seedlog.Infof("Grpc stream (server): %s", conn.Spec().Procedure)
		seedlog.Debugf("Grpc headers: %+v", conn.RequestHeader())
		seedErrorCode := uint32(0)
		err := next(ctx, &loggingStreamingHandlerConn{StreamingHandlerConn: conn})
		if err != nil {
			seedlog.Errorf("%s error: %v", conn.Spec().Procedure, err)
			seedErr := &seederr.SeedError{}
			if errors.As(err, &seedErr) {
				seedErrorCode = seedErr.Code()
				err = seedErr.Unwrap()
			}
		}
		grpcErrorCode := seedErrorCode & 0xff
		if grpcErrorCode != 0 {
			return connect.NewError(connect.Code(grpcErrorCode), err)
		}
		return err
	}
}

func NewLogInterceptor() *logInterceptor {
	return &logInterceptor{}
}
