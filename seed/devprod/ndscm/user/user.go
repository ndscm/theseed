package user

import "github.com/ndscm/theseed/seed/infra/flag/go/seedflag"

var flagUserHandle = seedflag.DefineString("user_handle", "", "the user handle")
var flagUserEmail = seedflag.DefineString("user_email", "", "the user email")
var flagUserDisplayName = seedflag.DefineString("user_display_name", "", "the user display name")

func CurrentUserHandle() string {
	return flagUserHandle.Get()
}

func CurrentUserEmail() string {
	return flagUserEmail.Get()
}

func CurrentUserDisplayName() string {
	return flagUserDisplayName.Get()
}
