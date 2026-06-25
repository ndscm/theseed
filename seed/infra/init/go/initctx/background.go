package initctx

import (
	"context"

	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
)

var flagBearer = seedflag.DefineString("bearer", "", "Bearer token for client context.")

func Background() context.Context {
	bearer := flagBearer.Get()
	ctx := context.Background()
	if bearer != "" {
		ctx = seedbearer.WithBearer(ctx, bearer)
	}
	return ctx
}
