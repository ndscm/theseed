package signedjwt

import (
	"strings"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

func WrapOpenidClient(
	discoveryUrl string,
	clientId string, clientSecret string,
	fallbackClient *openid.OpenidClient,
) (*openid.OpenidClient, error) {
	client := (*openid.OpenidClient)(nil)
	if clientId == "" {
		if fallbackClient != nil {
			client = fallbackClient
		}
	} else {
		client = openid.NewOpenidClient(
			discoveryUrl, clientId, clientSecret,
		)
		if clientSecret == "" {
			audience := strings.TrimSuffix(discoveryUrl, "/.well-known/openid-configuration")
			clientAssertion, err := WrapRsaTokenSource(clientId, audience)
			if err != nil || clientAssertion == nil {
				seedlog.Warnf("Signed assertion is not used. clientId=%s, err=%v", clientId, err)
			} else {
				client.SetClientAssertion(clientAssertion)
			}
		}
	}
	if client == nil {
		return nil, seederr.WrapErrorf("client is not configured")
	}
	return client, nil
}
