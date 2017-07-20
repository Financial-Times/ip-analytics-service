package config

import (
	binder "github.com/Financial-Times/email-platform-tools/config"
)

type Config struct {
	GOENV      string `json:"goenv"`
	RabbitHost string `json:"rabbithost"`
}

func NewConfig(path string) (Config, error) {
	var cfg Config
	if err := binder.Bind(path, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
