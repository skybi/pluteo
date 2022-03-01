package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config represents the application configuration structure
type Config struct{}

// LoadFromEnv loads a new configuration structure using environment variables and an optional .env file
func LoadFromEnv() (*Config, error) {
	// Load a .env file if it exists
	_ = godotenv.Overload()

	// Load a new configuration structure using environment variables
	config := new(Config)
	if err := envconfig.Process("sb", config); err != nil {
		return nil, err
	}
	return config, nil
}
