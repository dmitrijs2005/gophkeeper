// Package config handles configuration for the server component,
// including defaults, JSON overlay, and command-line flags.
package config

import "time"

// Config holds runtime settings for the GophKeeper server.
//
// Fields:
//   - EndpointAddrGRPC: bind address for the public gRPC endpoint.
//   - DatabaseDSN: PostgreSQL DSN (pgx).
//   - SecretKey: HMAC secret for signing JWTs (HS256). Do not use test defaults in prod.
//   - AccessTokenValidityDuration / RefreshTokenValidityDuration: token lifetimes.
//   - S3RootUser / S3RootPassword: credentials for the S3-compatible backend.
//   - S3Bucket / S3Region / S3BaseEndpoint: object storage settings.
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

// LoadDefaults populates Config with sensible development defaults.
// NOTE: These values are insecure for production and should be overridden.
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

// LoadConfig builds a Config by applying defaults, then overlaying values
// from an optional JSON file and finally from command-line flags.
func LoadConfig() *Config {
	cfg := &Config{}
	cfg.LoadDefaults()
	parseJson(cfg)
	parseFlags(cfg)
	return cfg
}
