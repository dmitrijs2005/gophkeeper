// Package config handles the configuration for the server component,
// including parsing environment variables and command-line flags.
package config

import "time"

func (c *Config) LoadDefaults() {
	c.DatabaseDSN = "postgres://postgres:postgres@postgres:5432/gophkeeper?sslmode=disable"
	c.EndpointAddrGRPC = ":50051"
	c.SecretKey = "secretKey"
	c.AccessTokenValidityDuration = 1 * time.Minute
	c.RefreshTokenValidityDuration = 3 * time.Minute
}

type Config struct {
	EndpointAddrGRPC             string
	DatabaseDSN                  string
	SecretKey                    string
	AccessTokenValidityDuration  time.Duration
	RefreshTokenValidityDuration time.Duration
}

func LoadConfig() *Config {
	config := &Config{}
	config.LoadDefaults()

	// parseJson(config)
	// parseFlags(config)
	// parseEnv(config)

	return config
}
