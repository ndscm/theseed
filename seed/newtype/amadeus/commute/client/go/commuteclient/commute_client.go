package commuteclient

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/filesystem/proto/simplefspb"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/grpclog"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/infra/terminal/proto/terminalpb"
	"github.com/ndscm/theseed/seed/newtype/amadeus/commute/proto/commutepb"
	"github.com/ndscm/theseed/seed/newtype/amadeus/commute/proto/commutepbconnect"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var flagAmadeusCommuteServiceServer = seedflag.DefineString("amadeus_commute_service_server", "http://127.0.0.1:2623", "Amadeus commute service server address")

type AmadeusCommuteClient struct {
	client commutepbconnect.AmadeusCommuteServiceClient
}

func (c *AmadeusCommuteClient) SendBrainInput(
	ctx context.Context,
	input *brainpb.BrainInput,
) error {
	_, err := c.client.SendBrainInput(ctx, connect.NewRequest(
		&commutepb.SendBrainInputRequest{
			BrainInput: input,
		}))
	return err
}

// StartTerminal opens a terminal on the agent's playpen. The caller sends a start
// frame first, then keystrokes and resizes, and receives terminal output until
// the shell exits. The terminal lives as long as the stream.
func (c *AmadeusCommuteClient) StartTerminal(
	ctx context.Context,
) *connect.BidiStreamForClient[terminalpb.TerminalInputFrame, terminalpb.TerminalOutputFrame] {
	return c.client.StartTerminal(ctx)
}

// The file system calls below reach the workstation's file system, one whole
// path at a time. See AmadeusCommuteService for what each of them means; they
// are the same calls, and the errors they fail with — NotFound for a path that
// is not there, PermissionDenied for one the person may not touch — are the
// agent's own, carried back untouched so that whoever is driving the editor at
// the other end of hooin hears what the file system said.

// GetUserHome reports the home directory on the agent's workstation.
func (c *AmadeusCommuteClient) GetUserHome(ctx context.Context) (string, error) {
	res, err := c.client.GetUserHome(ctx, connect.NewRequest(&emptypb.Empty{}))
	if err != nil {
		return "", err
	}
	return res.Msg.GetPath(), nil
}

// Stat reports what a path is.
func (c *AmadeusCommuteClient) Stat(
	ctx context.Context, path string,
) (*simplefspb.FileStat, error) {
	res, err := c.client.Stat(ctx, connect.NewRequest(&simplefspb.FilePath{
		Path: path,
	}))
	if err != nil {
		return nil, err
	}
	return res.Msg, nil
}

// ReadDirectory lists the names in a directory, and what each one is.
func (c *AmadeusCommuteClient) ReadDirectory(
	ctx context.Context, path string,
) (map[string]uint32, error) {
	res, err := c.client.ReadDirectory(ctx, connect.NewRequest(&simplefspb.FilePath{
		Path: path,
	}))
	if err != nil {
		return nil, err
	}
	return res.Msg.GetEntries(), nil
}

// CreateDirectory creates one directory whose parent already exists.
func (c *AmadeusCommuteClient) CreateDirectory(ctx context.Context, path string) error {
	_, err := c.client.CreateDirectory(ctx, connect.NewRequest(&simplefspb.FilePath{
		Path: path,
	}))
	return err
}

// ReadFile reads a whole file.
func (c *AmadeusCommuteClient) ReadFile(ctx context.Context, path string) ([]byte, error) {
	res, err := c.client.ReadFile(ctx, connect.NewRequest(&simplefspb.FilePath{
		Path: path,
	}))
	if err != nil {
		return nil, err
	}
	return res.Msg.GetContent(), nil
}

// WriteFile replaces a file's contents, creating it if create allows and
// replacing an existing one if overwrite does.
func (c *AmadeusCommuteClient) WriteFile(
	ctx context.Context, path string, content []byte, create bool, overwrite bool,
) error {
	_, err := c.client.WriteFile(ctx, connect.NewRequest(&simplefspb.WriteFileRequest{
		Path:      path,
		Content:   content,
		Create:    create,
		Overwrite: overwrite,
	}))
	return err
}

// Delete deletes a path, and a directory's contents with it when recursive
// says so.
func (c *AmadeusCommuteClient) Delete(
	ctx context.Context, path string, recursive bool,
) error {
	_, err := c.client.Delete(ctx, connect.NewRequest(&simplefspb.DeleteRequest{
		Path:      path,
		Recursive: recursive,
	}))
	return err
}

// Rename moves a path onto another, replacing what is there if overwrite
// allows it.
func (c *AmadeusCommuteClient) Rename(
	ctx context.Context, sourcePath string, destinationPath string, overwrite bool,
) error {
	_, err := c.client.Rename(ctx, connect.NewRequest(&simplefspb.RenameRequest{
		SourcePath:      sourcePath,
		DestinationPath: destinationPath,
		Overwrite:       overwrite,
	}))
	return err
}

type amadeusCommuteClientOptions struct {
	httpClient *http.Client
	server     string
}

type AmadeusCommuteClientOption func(*amadeusCommuteClientOptions)

func WithHttpClient(httpClient *http.Client) AmadeusCommuteClientOption {
	return func(opts *amadeusCommuteClientOptions) {
		opts.httpClient = httpClient
	}
}

func WithServer(server string) AmadeusCommuteClientOption {
	return func(opts *amadeusCommuteClientOptions) {
		opts.server = server
	}
}

func NewAmadeusCommuteClient(opts ...AmadeusCommuteClientOption) *AmadeusCommuteClient {
	o := &amadeusCommuteClientOptions{
		httpClient: http.DefaultClient,
		server:     flagAmadeusCommuteServiceServer.Get(),
	}
	for _, opt := range opts {
		opt(o)
	}
	client := commutepbconnect.NewAmadeusCommuteServiceClient(
		seedbearer.InterceptBearerTransport(o.httpClient),
		o.server,
		connect.WithInterceptors(grpclog.NewLogInterceptor()),
	)
	return &AmadeusCommuteClient{client}
}
