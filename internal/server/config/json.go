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
	EndpointAddrGRPC             string         `json:"endpoint_addr_grpc"`
	DatabaseDSN                  string         `json:"database_dsn"`
	SecretKey                    string         `json:"secret_key"`
	AccessTokenValidityDuration  timex.Duration `json:"access_token_validity_duration"`
	RefreshTokenValidityDuration timex.Duration `json:"refresh_token_validity_duration"`
	S3RootUser                   string         `json:"s3_root_user"`
	S3RootPassword               string         `json:"s3_root_password"`
	S3Bucket                     string         `json:"s3_bucket"`
	S3Region                     string         `json:"s3_region"`
	S3BaseEndpoint               string         `json:"s3_base_endpoint"`
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

	config.EndpointAddrGRPC = c.EndpointAddrGRPC
	config.DatabaseDSN = c.DatabaseDSN
	config.SecretKey = c.SecretKey
	config.AccessTokenValidityDuration = time.Duration(c.AccessTokenValidityDuration.Duration)
	config.RefreshTokenValidityDuration = time.Duration(c.RefreshTokenValidityDuration.Duration)
	config.S3RootUser = c.S3RootUser
	config.S3RootPassword = c.S3RootPassword
	config.S3Bucket = c.S3Bucket
	config.S3Region = c.S3Region
	config.S3BaseEndpoint = c.S3BaseEndpoint
}
