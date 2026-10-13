package vpcdirect

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagVpcDirect = seedflag.DefineStringList(
	"vpc_direct", []string{},
	"Redirect outbound requests for public hosts to vpc origins. ",
)

func ParseEntries(entries []string) (map[string]*url.URL, error) {
	directMap := map[string]*url.URL{}
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		requestHost, directOrigin, ok := strings.Cut(entry, "=")
		if !ok {
			return nil, seederr.WrapErrorf("invalid entry: %q", entry)
		}
		origin, err := url.Parse(strings.TrimSpace(directOrigin))
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		if origin.Scheme == "" || origin.Host == "" {
			return nil, seederr.WrapErrorf("direct origin must be an absolute URL: %q", directOrigin)
		}
		directMap[strings.TrimSpace(requestHost)] = origin
	}
	return directMap, nil
}

type vpcDirectTransport struct {
	directMap map[string]*url.URL

	next http.RoundTripper
}

func (t *vpcDirectTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	directUrl, ok := t.directMap[request.URL.Host]
	if !ok {
		return t.next.RoundTrip(request)
	}
	requestHost := request.URL.Host
	requestScheme := request.URL.Scheme
	clonedRequest := request.Clone(request.Context())
	clonedRequest.URL.Scheme = directUrl.Scheme
	clonedRequest.URL.Host = directUrl.Host
	clonedRequest.Host = directUrl.Host
	clonedRequest.Header.Set("X-Forwarded-Host", requestHost)
	clonedRequest.Header.Set("X-Forwarded-Proto", requestScheme)
	return t.next.RoundTrip(clonedRequest)
}

var _ http.RoundTripper = (*vpcDirectTransport)(nil)

func WrapVpcDirectTransport(directMap map[string]*url.URL, next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	return &vpcDirectTransport{
		directMap: directMap,

		next: next,
	}
}

type VpcDirectTransportOptions struct {
	Entries []string
}

func CreateVpcDirectTransport(opts *VpcDirectTransportOptions, next http.RoundTripper) (http.RoundTripper, error) {
	if opts == nil {
		opts = &VpcDirectTransportOptions{}
	}
	if len(opts.Entries) == 0 {
		opts.Entries = flagVpcDirect.Get()
	}
	directMap, err := ParseEntries(opts.Entries)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return WrapVpcDirectTransport(directMap, next), nil
}

func CreateVpcDirectClient() (*http.Client, error) {
	transport, err := CreateVpcDirectTransport(nil, nil)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return &http.Client{
		Transport: transport,
	}, nil
}
