package service

import (
	"context"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/cloud/login/go/keycloaklogin"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/filesystem/proto/simplefspb"
	"github.com/ndscm/theseed/seed/newtype/amadeus/commute/client/go/commuteclient"
	"github.com/ndscm/theseed/seed/newtype/hooin/onsite"
	"github.com/ndscm/theseed/seed/newtype/hooin/raid/proto/raidpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

// HooinRaidService reaches the file system of a person's workstation from
// outside it.
//
// The paths it reaches are held by nobody: the workstation has them, and this
// forwards to the agent on it. What it holds instead is the office, which is
// where the duty of the person whose workstation it is is found — and a duty is
// the only way to a workstation.
type HooinRaidService struct {
	office *onsite.Office
}

// connectFileSystem finds the file system of a person's workstation, for a
// caller who is allowed to reach into it.
//
// Every file system call goes through here, and each is checked on its own:
// unlike a terminal, which is checked when it is opened and then typed at
// through a session nobody else can name, a path is named afresh every time.
// There is nothing held between two calls to stand in for the check.
//
// The role checked is the one a terminal asks for — invading a workstation is
// invading it, whether by shell or by editor, and a person who may sit at the
// one may read what the other would have shown them. What their file system
// then allows is decided on the workstation, where the agent runs each call as
// them.
func (svc *HooinRaidService) connectFileSystem(
	ctx context.Context, personId string,
) (*commuteclient.AmadeusCommuteClient, error) {
	loginUser, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	if personId == "" {
		return nil, seederr.CodeErrorf(codes.InvalidArgument, "person_id is required")
	}
	err = keycloaklogin.VerifyRole(loginUser, "", "raid:"+personId)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	duty := svc.office.GetDuty(personId)
	if duty == nil {
		return nil, seederr.CodeErrorf(codes.FailedPrecondition,
			"person %q is not on duty", personId)
	}
	return duty.SimpleFileSystem(), nil
}

// GetUserHome reports where an editor opening this workstation starts.
func (svc *HooinRaidService) GetUserHome(
	ctx context.Context,
	req *connect.Request[raidpb.GetUserHomeRequest],
) (*connect.Response[simplefspb.FilePath], error) {
	fileSystem, err := svc.connectFileSystem(ctx, req.Msg.GetPersonId())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	path, err := fileSystem.GetUserHome(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	return connect.NewResponse(&simplefspb.FilePath{
		Path: path,
	}), nil
}

// Stat reports what a path on the workstation is.
func (svc *HooinRaidService) Stat(
	ctx context.Context,
	req *connect.Request[raidpb.StatRequest],
) (*connect.Response[simplefspb.FileStat], error) {
	fileSystem, err := svc.connectFileSystem(ctx, req.Msg.GetPersonId())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	// The stat the agent reports is the one the caller gets: what a path is does
	// not change on the way through hooin.
	stat, err := fileSystem.Stat(ctx, req.Msg.GetRequest().GetPath())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	return connect.NewResponse(stat), nil
}

// ReadDirectory lists a directory on the workstation.
func (svc *HooinRaidService) ReadDirectory(
	ctx context.Context,
	req *connect.Request[raidpb.ReadDirectoryRequest],
) (*connect.Response[simplefspb.ReadDirectoryResponse], error) {
	fileSystem, err := svc.connectFileSystem(ctx, req.Msg.GetPersonId())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	entries, err := fileSystem.ReadDirectory(ctx, req.Msg.GetRequest().GetPath())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	return connect.NewResponse(&simplefspb.ReadDirectoryResponse{
		Entries: entries,
	}), nil
}

// CreateDirectory creates one directory on the workstation.
func (svc *HooinRaidService) CreateDirectory(
	ctx context.Context,
	req *connect.Request[raidpb.CreateDirectoryRequest],
) (*connect.Response[emptypb.Empty], error) {
	fileSystem, err := svc.connectFileSystem(ctx, req.Msg.GetPersonId())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	err = fileSystem.CreateDirectory(ctx, req.Msg.GetRequest().GetPath())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// ReadFile reads a whole file off the workstation.
func (svc *HooinRaidService) ReadFile(
	ctx context.Context,
	req *connect.Request[raidpb.ReadFileRequest],
) (*connect.Response[simplefspb.FileContent], error) {
	fileSystem, err := svc.connectFileSystem(ctx, req.Msg.GetPersonId())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	content, err := fileSystem.ReadFile(ctx, req.Msg.GetRequest().GetPath())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	return connect.NewResponse(&simplefspb.FileContent{
		Content: content,
	}), nil
}

// WriteFile replaces a file on the workstation.
func (svc *HooinRaidService) WriteFile(
	ctx context.Context,
	req *connect.Request[raidpb.WriteFileRequest],
) (*connect.Response[emptypb.Empty], error) {
	fileSystem, err := svc.connectFileSystem(ctx, req.Msg.GetPersonId())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	request := req.Msg.GetRequest()
	err = fileSystem.WriteFile(ctx,
		request.GetPath(), request.GetContent(), request.GetCreate(), request.GetOverwrite())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// Delete deletes a path on the workstation.
func (svc *HooinRaidService) Delete(
	ctx context.Context,
	req *connect.Request[raidpb.DeleteRequest],
) (*connect.Response[emptypb.Empty], error) {
	fileSystem, err := svc.connectFileSystem(ctx, req.Msg.GetPersonId())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	request := req.Msg.GetRequest()
	err = fileSystem.Delete(ctx, request.GetPath(), request.GetRecursive())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// Rename moves a path on the workstation onto another.
func (svc *HooinRaidService) Rename(
	ctx context.Context,
	req *connect.Request[raidpb.RenameRequest],
) (*connect.Response[emptypb.Empty], error) {
	fileSystem, err := svc.connectFileSystem(ctx, req.Msg.GetPersonId())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	request := req.Msg.GetRequest()
	err = fileSystem.Rename(ctx,
		request.GetSourcePath(), request.GetDestinationPath(), request.GetOverwrite())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// NewHooinRaidService reaches workstations through the given office: it is where
// the duty of a person on one is found, and there is nothing to reach a
// workstation through but a duty.
func NewHooinRaidService(office *onsite.Office) *HooinRaidService {
	return &HooinRaidService{
		office: office,
	}
}
