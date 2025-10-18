package config

import "time"

// Config holds runtime settings for the GophKeeper CLI.
//
// Fields:
//   - ServerEndpointAddr: host:port of the backend gRPC endpoint.
//   - OnlineCheckInterval: how often the client probes server reachability.
//
// Units: OnlineCheckInterval is a time.Duration (e.g., 3*time.Second).
type Config struct {
	ServerEndpointAddr  string
	OnlineCheckInterval time.Duration
}

// LoadDefaults populates c with sensible defaults.
func (c *Config) LoadDefaults() {
	c.ServerEndpointAddr = "127.0.0.1:50051"
	c.OnlineCheckInterval = 3 * time.Second
}

// LoadConfig constructs a Config, applies defaults, then overlays values from
// JSON (if present) and command-line flags (if present). Later sources take
// precedence over earlier ones.
func LoadConfig() *Config {
	cfg := &Config{}
	cfg.LoadDefaults()
	parseJson(cfg)
	parseFlags(cfg)
	return cfg
}
