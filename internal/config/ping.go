package config

import "github.com/caarlos0/env/v8"

type Ping struct {
	Key     string `env:"PING_SERVICE_KEY" envDefault:"" json:"service_key,omitempty"`
	Address string `env:"PING_SERVICE_ADDRESS" envDefault:"" json:"service_address,omitempty"`
}

func BuildPing(cfg *Config) error {
	ping := &Ping{}
	if err := env.Parse(ping); err != nil {
		return err
	}

	cfg.Ping = *ping
	return nil
}
