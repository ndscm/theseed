package service

import (
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func BuildFromCelIdentName(e *exprpb.Expr) (string, error) {
	ident := e.GetIdentExpr()
	if ident == nil {
		return "", seederr.WrapErrorf("expected field identifier, got %T", e.ExprKind)
	}
	return ident.GetName(), nil
}

func BuildFromCelTimestamp(call *exprpb.Expr_Call) (any, error) {
	if call.GetFunction() != "timestamp" {
		return nil, seederr.WrapErrorf("unsupported function in value: %s", call.GetFunction())
	}
	args := call.GetArgs()
	if len(args) != 1 {
		return nil, seederr.WrapErrorf("timestamp() requires exactly 1 argument")
	}
	sv, ok := args[0].GetConstExpr().GetConstantKind().(*exprpb.Constant_StringValue)
	if !ok {
		return nil, seederr.WrapErrorf("timestamp() requires a string argument")
	}
	t, err := time.Parse(time.RFC3339, sv.StringValue)
	if err != nil {
		return nil, seederr.WrapErrorf("invalid timestamp %q: %w", sv.StringValue, err)
	}
	return t, nil
}

func BuildFromCelConstValue(e *exprpb.Expr) (any, error) {
	call := e.GetCallExpr()
	if call != nil {
		switch call.GetFunction() {
		case "timestamp":
			return BuildFromCelTimestamp(call)
		default:
			return nil, seederr.WrapErrorf("unsupported function in value: %s", call.GetFunction())
		}
	}
	c := e.GetConstExpr()
	if c == nil {
		return nil, seederr.WrapErrorf("expected constant value, got %T", e.ExprKind)
	}
	switch v := c.ConstantKind.(type) {
	case *exprpb.Constant_StringValue:
		return v.StringValue, nil
	case *exprpb.Constant_Int64Value:
		return v.Int64Value, nil
	case *exprpb.Constant_BoolValue:
		return v.BoolValue, nil
	case *exprpb.Constant_DoubleValue:
		return v.DoubleValue, nil
	default:
		return nil, seederr.WrapErrorf("unsupported constant type: %T", c.ConstantKind)
	}
}
