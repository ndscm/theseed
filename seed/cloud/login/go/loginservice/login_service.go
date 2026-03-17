package loginservice

import (
	"context"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/cloud/login/proto/loginpb"
)

type LoginService struct{}

func (svc *LoginService) GetLoginStatus(
	ctx context.Context,
	request *connect.Request[loginpb.GetLoginStatusRequest],
) (*connect.Response[loginpb.LoginStatus], error) {
	result := loginpb.LoginStatus{}
	loginUser, err := login.LoginUser(ctx)
	if err == nil && loginUser != nil {
		result.UserUuid = loginUser.Sub
		result.UserHandle = loginUser.PreferredUsername
		result.DisplayName = loginUser.Name
	}
	return connect.NewResponse(&result), nil
}
