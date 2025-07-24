package seedctx

import (
	"context"
	"os"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type MptContextKey string

func Performer(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", seederr.WrapErrorf("nil context provided")
	}
	performer, ok := ctx.Value(MptContextKey("performer")).(string)
	if !ok {
		return "", seederr.WrapErrorf("performer not found in context")
	}
	return performer, nil
}

func WithPerformer(parent context.Context, performer string) context.Context {
	return context.WithValue(parent, MptContextKey("performer"), performer)
}

func Background() context.Context {
	ndUserHandle := os.Getenv("ND_USER_HANDLE")
	if ndUserHandle == "" {
		panic("nd user handle is not set")
	}
	ctx := context.Background()
	ctx = WithPerformer(ctx, ndUserHandle+"@ndscm.com")
	return ctx
}
