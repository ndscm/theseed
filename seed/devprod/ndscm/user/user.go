package user

import (
	"regexp"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagUserHandle = seedflag.DefineString("user_handle", "", "the user handle")
var flagUserDomain = seedflag.DefineString("user_domain", "", "the user email domain")
var flagUserDisplayName = seedflag.DefineString("user_display_name", "", "the user display name")

var validHandle = regexp.MustCompile(`^[a-z][a-z0-9._-]*$`)
var validDomain = regexp.MustCompile(`^[a-z0-9]+([.-][a-z0-9]+)*$`)
var invalidDisplayName = regexp.MustCompile("[\"\\\\\\n\\r\\t]")

// CurrentUserHandle returns the user handle from the --user_handle flag.
// The handle must start with a lowercase letter and contain only lowercase
// letters, digits, dashes, underscores, and dots.
func CurrentUserHandle() (string, error) {
	handle := flagUserHandle.Get()
	if handle == "" {
		return "", seederr.WrapErrorf("user_handle is required")
	}
	if !validHandle.MatchString(handle) {
		return "", seederr.WrapErrorf("user_handle %q is invalid: only lowercase letters, digits, dash, underscore, and dot are allowed", handle)
	}
	return handle, nil
}

// CurrentUserDomain returns the user email domain from the --user_domain flag.
// The domain is one or more labels of lowercase letters and digits joined by
// single dots or dashes, with no leading, trailing, or consecutive separators.
// A single-label private domain such as "localhost" is accepted; a public
// top-level domain is not required.
func CurrentUserDomain() (string, error) {
	domain := flagUserDomain.Get()
	if domain == "" {
		return "", seederr.WrapErrorf("user_domain is required")
	}
	if !validDomain.MatchString(domain) {
		return "", seederr.WrapErrorf("user_domain %q is invalid: use lowercase letters and digits in dot- or dash-separated labels", domain)
	}
	return domain, nil
}

// CurrentUserEmail returns the user email constructed as handle@domain from the
// validated --user_handle and --user_domain flags. It returns an error if either
// flag is missing or invalid; see CurrentUserHandle and CurrentUserDomain.
func CurrentUserEmail() (string, error) {
	handle, err := CurrentUserHandle()
	if err != nil {
		return "", err
	}
	domain, err := CurrentUserDomain()
	if err != nil {
		return "", err
	}
	email := handle + "@" + domain
	return email, nil
}

// CurrentUserDisplayName returns the display name from the --user_display_name flag.
// The name must not contain quotes, backslashes, or control characters (newline,
// carriage return, tab) to stay safe in dotenv format.
func CurrentUserDisplayName() (string, error) {
	name := flagUserDisplayName.Get()
	if name == "" {
		return "", seederr.WrapErrorf("user_display_name is required")
	}
	if invalidDisplayName.MatchString(name) {
		return "", seederr.WrapErrorf("user_display_name %q is invalid: quotes, backslashes, and control characters are not allowed", name)
	}
	return name, nil
}
