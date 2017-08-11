package config

import (
	binder "github.com/Financial-Times/email-platform-tools/config"
)

// Config provides app wide config vars
type Config struct {
	GOENV              string `json:"goenv"`
	APIKey             string `json:"apikey"`
	RabbitHost         string `json:"rabbithost"`
	QueueName          string `json:"queuename"`
	Port               string `json:"port"`
	AWSRegion          string `json:"awsregion"`
	AWSAccessKeyID     string `json:"awsaccesskeyid"`
	AWSSecretAccessKey string `json:"awssecretaccesskey"`
	KinesisStream      string `json:"kinesisstream"`
}

// NewConfig returns a new Config instance bound with yaml file
func NewConfig(path string) (Config, error) {
	var cfg Config
	if err := binder.Bind(path, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
