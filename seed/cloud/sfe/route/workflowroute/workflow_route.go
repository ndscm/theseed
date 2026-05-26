package workflowroute

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagWorkflowServiceServer = seedflag.DefineString("workflow_service_server", "", "URL of Workflow service server")

func CreateWorkflowRoute(transport http.RoundTripper) (*httputil.ReverseProxy, error) {
	serverUrl, err := url.Parse(flagWorkflowServiceServer.Get())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	reverseProxy := &httputil.ReverseProxy{
		Transport: transport,
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(serverUrl)
			r.SetXForwarded()
		},
	}
	return reverseProxy, nil
}
