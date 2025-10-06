package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/flagx"
	"github.com/dmitrijs2005/gophkeeper/internal/timex"
)

// JsonConfig defines a configuration structure tailored for JSON unmarshalling.
// It uses common.Duration for interval fields, which allows parsing both
// string values such as "1s" and integer nanoseconds.
//
// This struct is an intermediate DTO (Data Transfer Object) used only for
// reading JSON configuration files. After unmarshalling, its fields are
// copied into the runtime Config struct which uses time.Duration.
type JsonConfig struct {
	ServerEndpointAddr  string         `json:"server_endpoint_addr"`
	OnlineCheckInterval timex.Duration `json:"online_check_interval"`
}

// parseJson loads configuration values from a JSON file into the provided
// Config instance.
//
// The lookup order for the JSON file path is:
//
//	The -c or -config command-line flags.
//	If it is not set, no JSON file is loaded.
//
// If the file path is found, parseJson attempts to read and unmarshal it
// into a JsonConfig. The resulting values are copied into the target Config.
// If the file cannot be read or contains invalid JSON, the function panics.
//
// Fields populated:
//   - ServerEndpointAddr
//   - OnlineCheckInterval
//
// The caller is expected to merge these values with defaults and command-line flags as part of the full configuration process.
func parseJson(config *Config) {

	// try flags
	jsonConfigFile := flagx.JsonConfigFlags()

	// nothing to load
	if jsonConfigFile == "" {
		return
	}

	c := &JsonConfig{}

	file, err := os.ReadFile(jsonConfigFile)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(file, c)
	if err != nil {
		panic(err)
	}

	config.ServerEndpointAddr = c.ServerEndpointAddr
	config.OnlineCheckInterval = time.Duration(c.OnlineCheckInterval.Duration)
}
