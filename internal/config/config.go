package config

import (
	"github.com/bugfixes/go-bugfixes/logs"
	ConfigBuilder "github.com/keloran/go-config"
)

type Config struct {
	Identity
	Ping
	ConfigBuilder.Config
}

func Build() (*Config, error) {
	cfg := &Config{}

	if err := BuildIdentity(cfg); err != nil {
		return cfg, logs.Errorf("buildidentity: %v", err)
	}

	if err := BuildPing(cfg); err != nil {
		return cfg, logs.Errorf("buildping: %v", err)
	}

	conf, err := ConfigBuilder.Build(ConfigBuilder.Vault, ConfigBuilder.Mongo)
	if err != nil {
		return cfg, logs.Errorf("configbuilder: %v", err)
	}
	cfg.Config = *conf

	return cfg, nil
}
