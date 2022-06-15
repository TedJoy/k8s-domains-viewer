package config

import (
	"github.com/caarlos0/env/v6"
)

func init() {
	MyEnvConfig = EnvConfig{}
	if err := env.Parse(&MyEnvConfig.Application); err != nil {
		panic(err)
	}

	if err := env.Parse(&MyEnvConfig.HTTPServer); err != nil {
		panic(err)
	}
}
