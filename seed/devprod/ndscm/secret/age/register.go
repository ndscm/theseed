package age

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/secret"
)

func init() {
	secret.Register(".age", &AgeProvider{})
}
