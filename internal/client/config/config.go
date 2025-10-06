// Package config handles initialization and validation of the agent configuration,
// including reading from environment variables and command-line flags.
package config

import "time"

func (c *Config) LoadDefaults() {
	c.ServerEndpointAddr = "127.0.0.1:50051"
	c.OnlineCheckInterval = 3 * time.Second
}

type Config struct {
	ServerEndpointAddr  string
	OnlineCheckInterval time.Duration
}

func LoadConfig() *Config {
	config := &Config{}
	config.LoadDefaults()
	parseJson(config)
	parseFlags(config)

	return config
}
