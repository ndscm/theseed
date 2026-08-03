package clientcore

import (
	"fmt"

	"github.com/ndscm/theseed/seed/devprod/ndscm/user"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type NdUserOptions struct {
	Args []string
}

func NdUser(options NdUserOptions) error {
	if len(options.Args) == 0 {
		return seederr.WrapErrorf("nd-user usage: nd user <get-handle|get-domain|get-email|get-display-name>")
	}
	subcommand := options.Args[0]
	if len(options.Args) > 1 {
		return seederr.WrapErrorf("nd-user %v takes no arguments", subcommand)
	}
	switch subcommand {
	case "get-handle":
		handle, err := user.CurrentUserHandle()
		if err != nil {
			return seederr.Wrap(err)
		}
		fmt.Printf("%s\n", handle)
	case "get-domain":
		domain, err := user.CurrentUserDomain()
		if err != nil {
			return seederr.Wrap(err)
		}
		fmt.Printf("%s\n", domain)
	case "get-email":
		email, err := user.CurrentUserEmail()
		if err != nil {
			return seederr.Wrap(err)
		}
		fmt.Printf("%s\n", email)
	case "get-display-name":
		displayName, err := user.CurrentUserDisplayName()
		if err != nil {
			return seederr.Wrap(err)
		}
		fmt.Printf("%s\n", displayName)
	default:
		return seederr.WrapErrorf("unknown nd-user subcommand %v", subcommand)
	}
	return nil
}
