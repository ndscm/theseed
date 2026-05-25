package certificateclient

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/cloud/sfe/certificate/proto/certificatepb"
	"github.com/ndscm/theseed/seed/cloud/sfe/certificate/proto/certificatepbconnect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/grpclog"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
)

var flagSfeCertificateServiceServer = seedflag.DefineString("sfe_certificate_service_server", "http://127.0.0.1:7332", "SFE certificate service server address")

type SfeCertificateClient struct {
	client certificatepbconnect.SfeCertificateServiceClient
}

func NewSfeCertificateClient(server string) *SfeCertificateClient {
	if server == "" {
		server = flagSfeCertificateServiceServer.Get()
	}
	client := certificatepbconnect.NewSfeCertificateServiceClient(
		seedbearer.InterceptBearerTransport(http.DefaultClient),
		server,
		connect.WithInterceptors(grpclog.NewLogInterceptor()),
	)
	return &SfeCertificateClient{client}
}

func (c *SfeCertificateClient) RenewCertificate(ctx context.Context, domain string) ([]byte, []byte, error) {
	resp, err := c.client.RenewCertificate(ctx, connect.NewRequest(&certificatepb.RenewCertificateRequest{
		Domain: domain,
	}))
	if err != nil {
		return nil, nil, seederr.Wrap(err)
	}
	return resp.Msg.Key, resp.Msg.Crt, nil
}
