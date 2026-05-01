package login

import (
	"context"

	"github.com/ndscm/theseed/seed/infra/auth/go/openidjwt"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func LoginUser(ctx context.Context) (*openidjwt.OpenidUserInfo, error) {
	openidUser, err := openidjwt.OpenidJwtUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return openidUser, nil
}

func EnsureLoginUser(ctx context.Context) (*openidjwt.OpenidUserInfo, error) {
	loginUser, err := LoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if loginUser == nil || loginUser.Sub == "" {
		return nil, seederr.WrapErrorf("user not logged in")
	}
	return loginUser, nil
}
