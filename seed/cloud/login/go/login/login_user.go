package login

import (
	"context"

	"github.com/ndscm/theseed/seed/infra/http/go/seedjwt"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func LoginUser(ctx context.Context) (*seedjwt.OpenidUserInfo, error) {
	jwtUser, err := seedjwt.JwtUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return jwtUser, nil
}

func EnsureLoginUser(ctx context.Context) (*seedjwt.OpenidUserInfo, error) {
	loginUser, err := LoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if loginUser == nil || loginUser.Sub == "" {
		return nil, seederr.WrapErrorf("user not logged in")
	}
	return loginUser, nil
}
