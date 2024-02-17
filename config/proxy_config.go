package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type ProxyConfig struct {
	UpstreamEndpoint   string `env:"UPSTREAM_ENDPOINT,required"`
	DenyResponseCode   int    `env:"DENY_RESPONSE_CODE" envDefault:"404"`
	ListenPort         int    `env:"PORT" envDefault:"3000"`
	ExitTimeoutSeconds int    `env:"EXIT_TIMEOUT" envDefault:"10"`
}

func LoadProxyConfig() (*ProxyConfig, error) {
	c := ProxyConfig{}
	if err := env.Parse(&c); err != nil {
		return nil, fmt.Errorf("load environment variables: %w", err)
	}
	return &c, nil
}
