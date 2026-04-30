package user

import (
	"fmt"
	"regexp"

	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagUserHandle = seedflag.DefineString("user_handle", "", "the user handle")
var flagUserEmail = seedflag.DefineString("user_email", "", "the user email")
var flagUserDisplayName = seedflag.DefineString("user_display_name", "", "the user display name")

var validHandle = regexp.MustCompile(`^[a-z][a-z0-9._-]*$`)
var validEmail = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
var invalidDisplayName = regexp.MustCompile("[\"\\\\\\n\\r\\t]")

// CurrentUserHandle returns the user handle from the --user_handle flag.
// The handle must start with a lowercase letter and contain only lowercase
// letters, digits, dashes, underscores, and dots.
func CurrentUserHandle() (string, error) {
	handle := flagUserHandle.Get()
	if handle == "" {
		return "", fmt.Errorf("user_handle is required")
	}
	if !validHandle.MatchString(handle) {
		return "", fmt.Errorf("user_handle %q is invalid: only lowercase letters, digits, dash, underscore, and dot are allowed", handle)
	}
	return handle, nil
}

// CurrentUserEmail returns the user email from the --user_email flag.
// The email must use only lowercase letters, digits, and common symbols
// in the local part, with a lowercase domain.
func CurrentUserEmail() (string, error) {
	email := flagUserEmail.Get()
	if email == "" {
		return "", fmt.Errorf("user_email is required")
	}
	if !validEmail.MatchString(email) {
		return "", fmt.Errorf("user_email %q is invalid", email)
	}
	return email, nil
}

// CurrentUserDisplayName returns the display name from the --user_display_name flag.
// The name must not contain quotes, backslashes, or control characters (newline,
// carriage return, tab) to stay safe in dotenv format.
func CurrentUserDisplayName() (string, error) {
	name := flagUserDisplayName.Get()
	if name == "" {
		return "", fmt.Errorf("user_display_name is required")
	}
	if invalidDisplayName.MatchString(name) {
		return "", fmt.Errorf("user_display_name %q is invalid: quotes, backslashes, and control characters are not allowed", name)
	}
	return name, nil
}
