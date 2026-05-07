package configloader

import (
	"encoding/json"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type StringOrStrings []string

func (s *StringOrStrings) UnmarshalJSON(data []byte) error {
	single := ""
	err := json.Unmarshal(data, &single)
	if err == nil {
		*s = []string{single}
		return nil
	}
	multi := []string{}
	err = json.Unmarshal(data, &multi)
	if err != nil {
		return seederr.Wrap(err)
	}
	*s = multi
	return nil
}
