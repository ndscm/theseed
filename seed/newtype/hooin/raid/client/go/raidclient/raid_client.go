package raidclient

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/filesystem/proto/simplefspb"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/grpclog"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/newtype/hooin/raid/proto/raidpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/raid/proto/raidpbconnect"
)

var flagHooinRaidServiceServer = seedflag.DefineString("hooin_raid_service_server", "http://127.0.0.1:4664", "Hooin raid service server address")

func HooinRaidServiceServerFlag() string {
	return flagHooinRaidServiceServer.Get()
}

type HooinRaidClient struct {
	client raidpbconnect.HooinRaidServiceClient
}

// GetUserHome reports the home directory on a person's workstation.
// See HooinRaidService.GetUserHome.
func (c *HooinRaidClient) GetUserHome(
	ctx context.Context, personId string,
) (string, error) {
	res, err := c.client.GetUserHome(ctx, connect.NewRequest(&raidpb.GetUserHomeRequest{
		PersonId: personId,
	}))
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return res.Msg.GetPath(), nil
}

// Stat reports what a path on a person's workstation is.
// See HooinRaidService.Stat.
func (c *HooinRaidClient) Stat(
	ctx context.Context, personId string, path string,
) (*simplefspb.FileStat, error) {
	res, err := c.client.Stat(ctx, connect.NewRequest(&raidpb.StatRequest{
		PersonId: personId,
		Request:  &simplefspb.FilePath{Path: path},
	}))
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return res.Msg, nil
}

// ReadDirectory lists a directory on a person's workstation.
// See HooinRaidService.ReadDirectory.
func (c *HooinRaidClient) ReadDirectory(
	ctx context.Context, personId string, path string,
) (map[string]uint32, error) {
	res, err := c.client.ReadDirectory(ctx, connect.NewRequest(&raidpb.ReadDirectoryRequest{
		PersonId: personId,
		Request:  &simplefspb.FilePath{Path: path},
	}))
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return res.Msg.GetEntries(), nil
}

// CreateDirectory creates one directory on a person's workstation.
// See HooinRaidService.CreateDirectory.
func (c *HooinRaidClient) CreateDirectory(
	ctx context.Context, personId string, path string,
) error {
	_, err := c.client.CreateDirectory(ctx, connect.NewRequest(&raidpb.CreateDirectoryRequest{
		PersonId: personId,
		Request:  &simplefspb.FilePath{Path: path},
	}))
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// ReadFile reads a whole file off a person's workstation.
// See HooinRaidService.ReadFile.
func (c *HooinRaidClient) ReadFile(
	ctx context.Context, personId string, path string,
) ([]byte, error) {
	res, err := c.client.ReadFile(ctx, connect.NewRequest(&raidpb.ReadFileRequest{
		PersonId: personId,
		Request:  &simplefspb.FilePath{Path: path},
	}))
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return res.Msg.GetContent(), nil
}

// WriteFile replaces a file on a person's workstation, creating it if create
// allows and replacing an existing one if overwrite does.
// See HooinRaidService.WriteFile.
func (c *HooinRaidClient) WriteFile(
	ctx context.Context, personId string, path string, content []byte, create bool, overwrite bool,
) error {
	_, err := c.client.WriteFile(ctx, connect.NewRequest(&raidpb.WriteFileRequest{
		PersonId: personId,
		Request: &simplefspb.WriteFileRequest{
			Path:      path,
			Content:   content,
			Create:    create,
			Overwrite: overwrite,
		},
	}))
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// Delete deletes a path on a person's workstation.
// See HooinRaidService.Delete.
func (c *HooinRaidClient) Delete(
	ctx context.Context, personId string, path string, recursive bool,
) error {
	_, err := c.client.Delete(ctx, connect.NewRequest(&raidpb.DeleteRequest{
		PersonId: personId,
		Request: &simplefspb.DeleteRequest{
			Path:      path,
			Recursive: recursive,
		},
	}))
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// Rename moves a path on a person's workstation onto another.
// See HooinRaidService.Rename.
func (c *HooinRaidClient) Rename(
	ctx context.Context, personId string, sourcePath string, destinationPath string, overwrite bool,
) error {
	_, err := c.client.Rename(ctx, connect.NewRequest(&raidpb.RenameRequest{
		PersonId: personId,
		Request: &simplefspb.RenameRequest{
			SourcePath:      sourcePath,
			DestinationPath: destinationPath,
			Overwrite:       overwrite,
		},
	}))
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

type hooinRaidClientOptions struct {
	httpClient *http.Client
	server     string
}

type HooinRaidClientOption func(*hooinRaidClientOptions)

func WithHttpClient(httpClient *http.Client) HooinRaidClientOption {
	return func(opts *hooinRaidClientOptions) {
		opts.httpClient = httpClient
	}
}

func WithServer(server string) HooinRaidClientOption {
	return func(opts *hooinRaidClientOptions) {
		opts.server = server
	}
}

func NewHooinRaidClient(opts ...HooinRaidClientOption) *HooinRaidClient {
	o := &hooinRaidClientOptions{
		httpClient: http.DefaultClient,
		server:     flagHooinRaidServiceServer.Get(),
	}
	for _, opt := range opts {
		opt(o)
	}
	client := raidpbconnect.NewHooinRaidServiceClient(
		seedbearer.InterceptBearerTransport(o.httpClient),
		o.server,
		connect.WithInterceptors(grpclog.NewLogInterceptor()),
	)
	return &HooinRaidClient{client}
}
