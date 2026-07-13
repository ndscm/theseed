package service

import (
	"context"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/filesystem/proto/simplefspb"
	"github.com/ndscm/theseed/seed/infra/terminal/proto/terminalpb"
	"github.com/ndscm/theseed/seed/newtype/amadeus/commute/proto/commutepb"
	"github.com/ndscm/theseed/seed/newtype/amadeus/onduty"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AmadeusCommuteService struct {
	conscious *onduty.Conscious
}

func (svc *AmadeusCommuteService) SendBrainInput(
	ctx context.Context,
	req *connect.Request[commutepb.SendBrainInputRequest],
) (*connect.Response[emptypb.Empty], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	// TODO(nagi): add fine-grained authorization
	brainInput := req.Msg.GetBrainInput()
	if brainInput == nil {
		return nil, seederr.CodeErrorf(codes.InvalidArgument, "brain_input is required")
	}
	err = svc.conscious.Input(brainInput)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// StartTerminal runs a terminal for the life of the stream. It is the only RPC
// here that needs the commute connection to carry a stream in both directions
// at once: the shell's output must reach the caller while the caller is still
// typing.
func (svc *AmadeusCommuteService) StartTerminal(
	ctx context.Context,
	stream *connect.BidiStream[terminalpb.TerminalInputFrame, terminalpb.TerminalOutputFrame],
) error {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}
	// TODO(nagi): add fine-grained authorization

	err = startTerminal(ctx, svc.conscious, stream)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// The workstation is the container, so there is no file system without one
// running, and a caller that asks for a path on a playpen that is not up is
// early rather than wrong.
func (svc *AmadeusCommuteService) GetUserHome(
	ctx context.Context,
	req *connect.Request[emptypb.Empty],
) (*connect.Response[simplefspb.FilePath], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	playpenController := svc.conscious.GetPlaypenController()
	if playpenController == nil {
		return nil, seederr.CodeErrorf(codes.FailedPrecondition, "no playpen container is running")
	}

	return connect.NewResponse(&simplefspb.FilePath{
		Path: playpenController.Home(),
	}), nil
}

func (svc *AmadeusCommuteService) Stat(
	ctx context.Context,
	req *connect.Request[simplefspb.FilePath],
) (*connect.Response[simplefspb.FileStat], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	playpenController := svc.conscious.GetPlaypenController()
	if playpenController == nil {
		return nil, seederr.CodeErrorf(codes.FailedPrecondition, "no playpen container is running")
	}
	fileSystem := playpenController.SimpleFileSystem()

	info, err := fileSystem.Stat(ctx, req.Msg.GetPath())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	// What the playpen says a path is is what the wire says it is, and what the
	// editor at the far end will read: one bitmask, VS Code's own, translated by
	// nobody.
	return connect.NewResponse(&simplefspb.FileStat{
		FileType: uint32(info.Type),
		Size:     uint64(info.Size),

		ModificationTimestampMs: info.ModificationTimestampMs,
		ChangeTimestampMs:       info.ChangeTimestampMs,
		CreationTimestampMs:     info.CreationTimestampMs,
	}), nil
}

func (svc *AmadeusCommuteService) ReadDirectory(
	ctx context.Context,
	req *connect.Request[simplefspb.FilePath],
) (*connect.Response[simplefspb.ReadDirectoryResponse], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	playpenController := svc.conscious.GetPlaypenController()
	if playpenController == nil {
		return nil, seederr.CodeErrorf(codes.FailedPrecondition, "no playpen container is running")
	}
	fileSystem := playpenController.SimpleFileSystem()

	entries, err := fileSystem.ReadDirectory(ctx, req.Msg.GetPath())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	entriesPb := map[string]uint32{}
	for name, fileType := range entries {
		entriesPb[name] = uint32(fileType)
	}

	return connect.NewResponse(&simplefspb.ReadDirectoryResponse{
		Entries: entriesPb,
	}), nil
}

func (svc *AmadeusCommuteService) CreateDirectory(
	ctx context.Context,
	req *connect.Request[simplefspb.FilePath],
) (*connect.Response[emptypb.Empty], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	playpenController := svc.conscious.GetPlaypenController()
	if playpenController == nil {
		return nil, seederr.CodeErrorf(codes.FailedPrecondition, "no playpen container is running")
	}
	fileSystem := playpenController.SimpleFileSystem()

	err = fileSystem.CreateDirectory(ctx, req.Msg.GetPath())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (svc *AmadeusCommuteService) ReadFile(
	ctx context.Context,
	req *connect.Request[simplefspb.FilePath],
) (*connect.Response[simplefspb.FileContent], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	playpenController := svc.conscious.GetPlaypenController()
	if playpenController == nil {
		return nil, seederr.CodeErrorf(codes.FailedPrecondition, "no playpen container is running")
	}
	fileSystem := playpenController.SimpleFileSystem()

	content, err := fileSystem.ReadFile(ctx, req.Msg.GetPath())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	return connect.NewResponse(&simplefspb.FileContent{
		Content: content,
	}), nil
}

func (svc *AmadeusCommuteService) WriteFile(
	ctx context.Context,
	req *connect.Request[simplefspb.WriteFileRequest],
) (*connect.Response[emptypb.Empty], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	playpenController := svc.conscious.GetPlaypenController()
	if playpenController == nil {
		return nil, seederr.CodeErrorf(codes.FailedPrecondition, "no playpen container is running")
	}
	fileSystem := playpenController.SimpleFileSystem()

	err = fileSystem.WriteFile(ctx,
		req.Msg.GetPath(), req.Msg.GetContent(), req.Msg.GetCreate(), req.Msg.GetOverwrite())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (svc *AmadeusCommuteService) Delete(
	ctx context.Context,
	req *connect.Request[simplefspb.DeleteRequest],
) (*connect.Response[emptypb.Empty], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	playpenController := svc.conscious.GetPlaypenController()
	if playpenController == nil {
		return nil, seederr.CodeErrorf(codes.FailedPrecondition, "no playpen container is running")
	}
	fileSystem := playpenController.SimpleFileSystem()

	err = fileSystem.Delete(ctx, req.Msg.GetPath(), req.Msg.GetRecursive())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (svc *AmadeusCommuteService) Rename(
	ctx context.Context,
	req *connect.Request[simplefspb.RenameRequest],
) (*connect.Response[emptypb.Empty], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	playpenController := svc.conscious.GetPlaypenController()
	if playpenController == nil {
		return nil, seederr.CodeErrorf(codes.FailedPrecondition, "no playpen container is running")
	}
	fileSystem := playpenController.SimpleFileSystem()

	err = fileSystem.Rename(ctx,
		req.Msg.GetSourcePath(), req.Msg.GetDestinationPath(), req.Msg.GetOverwrite())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func NewAmadeusCommuteService(conscious *onduty.Conscious) *AmadeusCommuteService {
	return &AmadeusCommuteService{
		conscious: conscious,
	}
}
