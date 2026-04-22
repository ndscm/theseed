package claudecli

import (
	"github.com/ndscm/theseed/seed/newtype/amadeus/brain"
)

func init() {
	b := NewClaudeCliBrain()
	brain.Register("claudecli", b)
}
