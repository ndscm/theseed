package keycloaklogin

import (
	"encoding/json/v2"
	"net/url"
	"slices"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

// Permission is one entry of the authorization claim carried by a Keycloak RPT
// (Requesting Party Token). It names a protected resource and the scopes the
// subject was granted on it. In our naming a scope is the action ("invade",
// "raid") and the resource is the thing acted on (a person's workstation uuid),
// so the pair "invade:daru-uuid" is scope "invade" on resource "daru-uuid".
type Permission struct {
	ResourceId   string   `json:"rsid,omitempty"`
	ResourceName string   `json:"rsname,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
}

// Authorization is the "authorization" claim of an RPT. Unlike resource_access,
// which Keycloak bakes into every login token, this claim is present only in a
// token minted through the UMA-ticket grant: a normal access token has no
// "authorization" claim at all. The gateway performs that grant and forwards the
// RPT, so the token reaching us carries this claim.
type Authorization struct {
	Permissions []Permission `json:"permissions,omitempty"`
}

// ResourceMatcher reports whether a resource named in a granted permission is
// the one being checked. It receives the rsname carried by the RPT and weighs it
// whole; a matcher that would rather weigh the parsed URL — the path alone, say,
// and not the authority that names the server it is on — is better written as a
// ResourceUrlMatcher and passed to VerifyResourceUrlPermission.
type ResourceMatcher func(resourceName string) bool

// ResourceUrlMatcher reports whether a resource named in a granted permission is
// the one being checked, weighing it as a parsed URL rather than a raw string.
// It lets a caller match on part of the name — the path, and not the authority
// (host) Keycloak registered the resource under — without parsing the rsname
// itself: VerifyResourceUrlPermission does the parsing.
type ResourceUrlMatcher func(resourceUrl *url.URL) bool

// ScopeMatcher reports whether a scope named in a granted permission is the one
// being checked.
type ScopeMatcher func(scope string) bool

// VerifyResourcePermission checks that the token granted the subject a scope
// scopeMatcher accepts on a resource resourceMatcher accepts. It is the
// fine-grained replacement for VerifyRole: where VerifyRole asked whether a
// client role named after the pair was present, this asks whether Keycloak's
// authorization engine granted the pair, and the grant travels in the RPT's
// authorization claim rather than in resource_access.
//
// The pair is given as matchers rather than as two strings so a caller may
// accept a family of grants without naming each.
//
// The openidUser must have been parsed from an RPT. If it was parsed from a
// plain access token the authorization claim is absent and every check fails;
// the gateway performs the UMA-ticket exchange and forwards the RPT, so the
// token reaching a service already carries the claim.
func VerifyResourcePermission(
	openidUser *openid.OpenidUserInfo, resourceMatcher ResourceMatcher, scopeMatcher ScopeMatcher,
) error {
	rawAuthorization, ok := openidUser.Inline["authorization"]
	if !ok {
		return seederr.WrapErrorf("no authorization claim for user %s (%s); token is not an RPT", openidUser.PreferredUsername, openidUser.Sub)
	}
	authorization := Authorization{}
	err := json.Unmarshal(rawAuthorization, &authorization)
	if err != nil {
		return seederr.Wrap(err)
	}

	granted := false
	for _, permission := range authorization.Permissions {
		if resourceMatcher(permission.ResourceName) && slices.ContainsFunc(permission.Scopes, scopeMatcher) {
			granted = true
			break
		}
	}
	if !granted {
		return seederr.WrapErrorf("no matching permission for user %s (%s)", openidUser.PreferredUsername, openidUser.Sub)
	}
	return nil
}

// VerifyResourceUrlPermission is VerifyResourcePermission for a caller that
// weighs the resource as a URL. It parses each granted permission's rsname and
// hands the URL to resourceUrlMatcher, so a matcher that cares only about the
// path — and not the authority (host) Keycloak registered the resource under —
// need not parse the name itself. An rsname that does not parse as a URL matches
// nothing.
func VerifyResourceUrlPermission(
	openidUser *openid.OpenidUserInfo, resourceUrlMatcher ResourceUrlMatcher, scopeMatcher ScopeMatcher,
) error {
	err := VerifyResourcePermission(openidUser, func(resourceName string) bool {
		resourceUrl, err := url.Parse(resourceName)
		if err != nil {
			return false
		}
		return resourceUrlMatcher(resourceUrl)
	}, scopeMatcher)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
