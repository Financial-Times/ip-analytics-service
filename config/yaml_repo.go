package config

import (
	configLoader "github.com/financial-times/email-platform-tools/config"
)

type YamlRepo struct {
	Path string
}

func (c *YamlRepo) GetAddress() (RabbitAddress, error) {
	var cfg Config

	if err := configLoader.Bind(r.Path, &cfg); err != nil {
		return RabbitAddress{}, err
	}
	return connectionConfigToAddress(cfg.Connection)
}
