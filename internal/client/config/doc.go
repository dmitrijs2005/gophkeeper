// Package config loads runtime configuration for the GophKeeper CLI.
//
// Sources & precedence
//
//  1. Built-in defaults (see (*Config).LoadDefaults).
//  2. Optional JSON file (see parseJson) selected via flags: -c or -config.
//  3. Command-line flags (see parseFlags), which override earlier values.
//
// Supported flags
//
//	-a string   address:port of the backend gRPC endpoint
//	-i int      online status check interval (seconds)
//
// # JSON schema
//
// The JSON loader uses timex.Duration for intervals, so values can be either
// strings like "3s" or integer nanoseconds:
//
//	{
//	  "server_endpoint_addr": "127.0.0.1:50051",
//	  "online_check_interval": "3s"
//	}
//
// Primary API
//
//   - type Config                     — holds ServerEndpointAddr and OnlineCheckInterval
//   - func LoadConfig() *Config       — builds Config by applying defaults, JSON, then flags
//   - func (*Config) LoadDefaults()   — sets sensible defaults
//
// Note: This package does not read environment variables directly; use the
// JSON file or flags to configure values.
package config
