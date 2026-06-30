package jwtcore

import "encoding/json"

// See RFC 7519 for JWT claims structure
// https://datatracker.ietf.org/doc/html/rfc7519
type JwtPayload struct {
	Iss string `json:"iss"`

	Aud json.RawMessage `json:"aud"`

	Exp *int64 `json:"exp"`

	Nbf *int64 `json:"nbf"`
}

func (p *JwtPayload) AudiencesContains(target string) bool {
	single := ""
	if json.Unmarshal(p.Aud, &single) == nil {
		return single == target
	}
	list := []string{}
	if json.Unmarshal(p.Aud, &list) == nil {
		for _, a := range list {
			if a == target {
				return true
			}
		}
	}
	return false
}
