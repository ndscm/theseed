package common

import (
	"github.com/joho/godotenv"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type NdConfig struct {
	MountHome string // Reserved for single mount point
}

func LoadConfig() (*NdConfig, error) {
	if false {
		err := godotenv.Load()
		if err != nil {
			return nil, seederr.WrapErrorf("%w", err)
		}
	}
	ndConfig := &NdConfig{
		MountHome: "",
	}
	return ndConfig, nil
}
