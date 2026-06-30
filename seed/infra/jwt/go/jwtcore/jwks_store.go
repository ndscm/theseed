package jwtcore

import (
	"crypto"
)

type JwksStore interface {
	GetByKid(issuer string, kid string) (crypto.PublicKey, error)
}
