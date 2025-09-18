// Package config handles initialization and validation of the agent configuration,
// including reading from environment variables and command-line flags.
package config

func (c *Config) LoadDefaults() {
	c.UseGRPC = true
	c.ServerEndpointAddr = ":50051"
}

type Config struct {
	ServerEndpointAddr string
	UseGRPC            bool
	// Key            string
	// ReportInterval time.Duration
	// PollInterval   time.Duration
	// SendRateLimit  int
	// CryptoKey      string
}

func LoadConfig() *Config {
	config := &Config{}
	config.LoadDefaults()

	// parseJson(config)
	// parseFlags(config)
	// parseEnv(config)

	return config
}
