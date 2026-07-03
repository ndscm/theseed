package servicetier

import (
	"net"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func cutLast(hostname string) (string, string, bool) {
	lastDot := strings.LastIndex(hostname, ".")
	if lastDot == -1 {
		return "", hostname, false
	}
	return hostname[:lastDot], hostname[lastDot+1:], true
}

type ServiceTier struct {
	RootDomain string
	Tenant     string
	Cluster    string
	Role       string
	Owner      string
	Tier       string
	Service    string

	Port string
}

func (t *ServiceTier) String() string {
	return t.Service + "." + t.Tier + "-" + t.Owner + "." + t.Role + "." + t.Cluster +
		"." + t.Tenant + "." + t.RootDomain + ":" + t.Port
}

type DomainVerifier interface {
	Verify(domain string) error
}

type ServiceTierParser struct {
	rootDomainVerifier   DomainVerifier
	tenantDomainVerifier DomainVerifier
}

func (p *ServiceTierParser) Parse(hostport string) (*ServiceTier, error) {
	hostname, port, err := net.SplitHostPort(hostport)
	if err != nil {
		hostname = hostport
		port = ""
	}

	// An IP address or a single-label host has no domain structure to parse.
	if net.ParseIP(hostname) != nil || !strings.Contains(hostname, ".") {
		return &ServiceTier{
			RootDomain: hostname,

			Port: port,
		}, nil
	}

	rootDomain := ""
	tenant := ""
	hostname, tld, _ := cutLast(hostname)
	hostname, sld, _ := cutLast(hostname)
	candidate := sld + "." + tld
	for {
		if p.rootDomainVerifier != nil {
			rootErr := p.rootDomainVerifier.Verify(candidate)
			if rootErr == nil {
				rootDomain = candidate
				break
			}
		}
		if p.tenantDomainVerifier != nil {
			tenantErr := p.tenantDomainVerifier.Verify(candidate)
			if tenantErr == nil {
				rootDomain = candidate
				hostname, tenant, _ = cutLast(hostname)
				break
			}
		}
		if hostname == "" {
			break
		}
		sub := ""
		hostname, sub, _ = cutLast(hostname)
		candidate = sub + "." + candidate
	}
	if rootDomain == "" {
		return nil, seederr.WrapErrorf("unknown root domain. host=%s", hostport)
	}

	hostParts := []string{}
	if hostname != "" {
		hostParts = strings.Split(hostname, ".")
	}
	cluster := ""
	role := ""
	owner := ""
	tier := ""
	service := ""
	switch len(hostParts) {
	case 0:
		break
	case 1:
		service = hostParts[0]
	case 2:
		service = hostParts[0]
		tier, owner, _ = strings.Cut(hostParts[1], "-")
	case 3:
		service = hostParts[0]
		tier, owner, _ = strings.Cut(hostParts[1], "-")
		cluster = hostParts[2]
	case 4:
		service = hostParts[0]
		tier, owner, _ = strings.Cut(hostParts[1], "-")
		role = hostParts[2]
		cluster = hostParts[3]
	default:
		return nil, seederr.WrapErrorf("too many host parts. host=%s", hostport)
	}
	return &ServiceTier{
		RootDomain: rootDomain,
		Tenant:     tenant,
		Cluster:    cluster,
		Role:       role,
		Owner:      owner,
		Tier:       tier,
		Service:    service,

		Port: port,
	}, nil
}

func NewServiceTierParser(rootDomainVerifier DomainVerifier, tenantDomainVerifier DomainVerifier) *ServiceTierParser {
	return &ServiceTierParser{
		rootDomainVerifier:   rootDomainVerifier,
		tenantDomainVerifier: tenantDomainVerifier,
	}
}

type StaticDomainVerifier struct {
	domains map[string]struct{}
}

func (v *StaticDomainVerifier) Verify(domain string) error {
	_, ok := v.domains[domain]
	if !ok {
		return seederr.WrapErrorf("invalid domain: %s", domain)
	}
	return nil
}

func NewStaticParser(rootDomains []string) *ServiceTierParser {
	rootDomainMap := make(map[string]struct{})
	for _, domain := range rootDomains {
		rootDomainMap[domain] = struct{}{}
	}
	return NewServiceTierParser(
		&StaticDomainVerifier{domains: rootDomainMap},
		nil,
	)
}

func NewStaticMultiTenantParser(rootDomains []string, tenantDomains []string) *ServiceTierParser {
	rootDomainMap := make(map[string]struct{})
	for _, domain := range rootDomains {
		rootDomainMap[domain] = struct{}{}
	}
	tenantDomainMap := make(map[string]struct{})
	for _, domain := range tenantDomains {
		tenantDomainMap[domain] = struct{}{}
	}
	return NewServiceTierParser(
		&StaticDomainVerifier{domains: rootDomainMap},
		&StaticDomainVerifier{domains: tenantDomainMap},
	)
}
