package service

import (
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/cel-go/cel"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// celBrainStepFilterEnv is the CEL environment for BrainStep list filters
// (AIP-160). The declared variables mirror the filterable BrainStep
// columns and their names match the database column names so the
// predicate walker can use identifiers directly.
var celBrainStepFilterEnv *cel.Env

func init() {
	celEnv, err := cel.NewEnv(
		cel.Variable("person_id", cel.StringType),
		cel.Variable("emit_time", cel.TimestampType),
		cel.Variable("type", cel.StringType),
		cel.Variable("topic", cel.StringType),
		cel.Variable("thread_uuid", cel.StringType),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create BrainStep CEL environment: %v", err))
	}
	celBrainStepFilterEnv = celEnv
}

func buildBrainStepPredicateFromCelLogical(
	call *exprpb.Expr_Call,
	combiner func(...*sql.Predicate) *sql.Predicate,
) (*sql.Predicate, error) {
	args := call.GetArgs()
	predicates := make([]*sql.Predicate, 0, len(args))
	for _, arg := range args {
		p, err := buildBrainStepPredicateFromCelExpr(arg)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, p)
	}
	return combiner(predicates...), nil
}

func buildBrainStepPredicateFromCelExpr(e *exprpb.Expr) (*sql.Predicate, error) {
	switch e.ExprKind.(type) {
	case *exprpb.Expr_CallExpr:
		return buildBrainStepPredicateFromCelCall(e.GetCallExpr())
	case *exprpb.Expr_IdentExpr:
		// A bare identifier is not meaningful for BrainStep since it has
		// no boolean columns; reject it rather than guess.
		return nil, seederr.WrapErrorf("unsupported bare identifier: %s", e.GetIdentExpr().GetName())
	default:
		return nil, seederr.WrapErrorf("unsupported expression: %T", e.ExprKind)
	}
}

func buildBrainStepPredicateFromCelComparison(
	call *exprpb.Expr_Call, op func(field string, value any) *sql.Predicate,
) (*sql.Predicate, error) {
	args := call.GetArgs()
	if len(args) != 2 {
		return nil, seederr.WrapErrorf("comparison requires exactly 2 arguments")
	}

	fieldName, err := BuildFromCelIdentName(args[0])
	if err != nil {
		return nil, err
	}
	value, err := BuildFromCelConstValue(args[1])
	if err != nil {
		return nil, err
	}

	return op(fieldName, value), nil
}

func buildBrainStepPredicateFromStringMethod(
	call *exprpb.Expr_Call, op func(field string, value string) *sql.Predicate,
) (*sql.Predicate, error) {
	target := call.GetTarget()
	if target == nil || len(call.GetArgs()) != 1 {
		return nil, seederr.WrapErrorf("%s requires a receiver and one argument", call.GetFunction())
	}

	fieldName, err := BuildFromCelIdentName(target)
	if err != nil {
		return nil, err
	}
	sv, ok := call.GetArgs()[0].GetConstExpr().GetConstantKind().(*exprpb.Constant_StringValue)
	if !ok {
		return nil, seederr.WrapErrorf("expected string constant for %s", call.GetFunction())
	}

	return op(fieldName, sv.StringValue), nil
}

func buildBrainStepPredicateFromCelCall(call *exprpb.Expr_Call) (*sql.Predicate, error) {
	switch call.GetFunction() {
	case "_&&_":
		return buildBrainStepPredicateFromCelLogical(call, sql.And)
	case "_||_":
		return buildBrainStepPredicateFromCelLogical(call, sql.Or)
	case "!_":
		inner, err := buildBrainStepPredicateFromCelExpr(call.GetArgs()[0])
		if err != nil {
			return nil, err
		}
		return sql.Not(inner), nil
	case "_==_":
		return buildBrainStepPredicateFromCelComparison(call, sql.EQ)
	case "_!=_":
		return buildBrainStepPredicateFromCelComparison(call, sql.NEQ)
	case "_<_":
		return buildBrainStepPredicateFromCelComparison(call, sql.LT)
	case "_<=_":
		return buildBrainStepPredicateFromCelComparison(call, sql.LTE)
	case "_>_":
		return buildBrainStepPredicateFromCelComparison(call, sql.GT)
	case "_>=_":
		return buildBrainStepPredicateFromCelComparison(call, sql.GTE)
	case "contains":
		return buildBrainStepPredicateFromStringMethod(call, sql.ContainsFold)
	case "startsWith":
		return buildBrainStepPredicateFromStringMethod(call, sql.HasPrefix)
	case "endsWith":
		return buildBrainStepPredicateFromStringMethod(call, sql.HasSuffix)
	default:
		return nil, seederr.WrapErrorf("unsupported function: %s", call.GetFunction())
	}
}

// compileBrainStepCelFilter parses a CEL filter string and returns an ent
// predicate. It returns nil when celFilter is empty.
func compileBrainStepCelFilter(celFilter string) (*sql.Predicate, error) {
	if celFilter == "" {
		return nil, nil
	}

	ast, issues := celBrainStepFilterEnv.Parse(celFilter)
	if issues != nil && issues.Err() != nil {
		return nil, seederr.WrapErrorf("invalid filter: %w", issues.Err())
	}

	checkedAst, issues := celBrainStepFilterEnv.Check(ast)
	if issues != nil && issues.Err() != nil {
		return nil, seederr.WrapErrorf("invalid filter: %w", issues.Err())
	}

	checkedExpr, err := cel.AstToCheckedExpr(checkedAst)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	return buildBrainStepPredicateFromCelExpr(checkedExpr.GetExpr())
}
