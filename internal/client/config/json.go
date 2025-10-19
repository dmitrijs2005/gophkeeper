package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/flagx"
	"github.com/dmitrijs2005/gophkeeper/internal/timex"
)

// JsonConfig is a DTO used exclusively for JSON unmarshalling.
// It relies on timex.Duration so JSON can specify intervals either as
// strings like "3s" or as integer nanoseconds. After parsing, values
// are copied into the runtime Config (which uses time.Duration).
type JsonConfig struct {
	ServerEndpointAddr  string         `json:"server_endpoint_addr"`
	OnlineCheckInterval timex.Duration `json:"online_check_interval"`
}

// parseJson overlays Config with values loaded from a JSON file.
//
// Lookup order for the JSON file path:
//  1. Command-line flags (-c or -config) via flagx.JsonConfigFlags().
//  2. If empty, no JSON is loaded and the function returns.
//
// Behavior:
//   - Reads and unmarshals the JSON into JsonConfig.
//   - Copies known fields into the provided Config.
//   - Panics on read or unmarshal errors (caller should recover if desired).
//
// Populated fields:
//   - ServerEndpointAddr
//   - OnlineCheckInterval
//
// Intended usage is: defaults -> parseJson -> parseFlags, where later stages
// override earlier ones.
func parseJson(cfg *Config) {
	// Resolve file path from flags.
	jsonConfigFile := flagx.JsonConfigFlags()
	if jsonConfigFile == "" {
		return
	}

	var jc JsonConfig

	data, err := os.ReadFile(jsonConfigFile)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, &jc); err != nil {
		panic(err)
	}

	cfg.ServerEndpointAddr = jc.ServerEndpointAddr
	cfg.OnlineCheckInterval = time.Duration(jc.OnlineCheckInterval.Duration)
}
