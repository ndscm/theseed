package tson

import (
	"encoding/json"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/tson/go/generator/genjson"
	"github.com/ndscm/theseed/seed/infra/tson/go/tsonparser"
)

func Unmarshal(src []byte, v any) error {
	node, err := tsonparser.Parse(src)
	if err != nil {
		return seederr.Wrap(err)
	}
	raw, err := genjson.Generate(node)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = json.Unmarshal(raw, v)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
