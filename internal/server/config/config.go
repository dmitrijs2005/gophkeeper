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
	c.S3RootUser = "admin"
	c.S3RootPassword = "secretpassword"
	c.S3Bucket = "vault"
	c.S3Region = "us-east-1"
	c.S3BaseEndpoint = "http://127.0.0.1:9000/"
}

type Config struct {
	EndpointAddrGRPC             string
	DatabaseDSN                  string
	SecretKey                    string
	AccessTokenValidityDuration  time.Duration
	RefreshTokenValidityDuration time.Duration
	S3RootUser                   string
	S3RootPassword               string
	S3Bucket                     string
	S3Region                     string
	S3BaseEndpoint               string
}

func LoadConfig() *Config {
	config := &Config{}
	config.LoadDefaults()
	parseJson(config)
	parseFlags(config)

	return config
}
