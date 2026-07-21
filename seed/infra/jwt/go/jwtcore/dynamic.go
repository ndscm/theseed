package jwtcore

import (
	"encoding/json/v2"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type StringOrStrings []string

func (s StringOrStrings) MarshalJSON() ([]byte, error) {
	if len(s) == 1 {
		data, err := json.Marshal(s[0])
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		return data, nil
	}
	data, err := json.Marshal([]string(s))
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return data, nil
}

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
