package git

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
)

func init() {
	scm.Register("git", &GitProvider{})
}
