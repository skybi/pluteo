package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"strings"
)

// Config represents the application configuration structure
type Config struct {
	Environment string `default:"prod"`

	PortalAPIListenAddress string `default:":8081" split_words:"true"`
	PortalAPIBaseAddress   string `default:"http://localhost:8081" split_words:"true"`

	OIDCProviderURL  string `split_words:"true"`
	OIDCClientID     string `split_words:"true"`
	OIDCClientSecret string `split_words:"true"`
}

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

// IsEnvProduction returns whether the application runs in production environment
func (config *Config) IsEnvProduction() bool {
	return strings.ToLower(config.Environment) != "dev"
}

// IsPortalAPISecure returns whether the portal API uses SSL in the end
func (config *Config) IsPortalAPISecure() bool {
	return strings.HasPrefix(strings.ToLower(config.PortalAPIBaseAddress), "https")
}
