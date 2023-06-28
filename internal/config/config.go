package config

import "github.com/caarlos0/env/v8"

type Config struct {
	Local
	Vault
	Mongo
	Identity
	Ping
}

func Build() (*Config, error) {
	cfg := &Config{}

	if err := BuildLocal(cfg); err != nil {
		return cfg, err
	}

	if err := BuildVault(cfg); err != nil {
		return cfg, err
	}

	if err := BuildMongo(cfg); err != nil {
		return cfg, err
	}

	if err := BuildIdentity(cfg); err != nil {
		return cfg, err
	}

	if err := BuildPing(cfg); err != nil {
		return cfg, err
	}

	if err := env.Parse(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
