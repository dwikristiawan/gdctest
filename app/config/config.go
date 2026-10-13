package config

import (
	"gdctest/pkg/database"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Server struct {
	Domain string `envconfig:"SERVER_DOMAIN" default:":8080"`
}

type Config struct {
	Server   Server          `envconfig:"SERVER"`
	Database database.Config `envconfig:"DB"`
}

func New(envPath string) (*Config, error) {
	if envPath != "" {
		_ = godotenv.Load(envPath)
	}
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
