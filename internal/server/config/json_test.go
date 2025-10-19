package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTempJSON(t *testing.T, dir, name string, data map[string]any) string {
	t.Helper()
	if dir == "" {
		dir = t.TempDir()
	}
	if name == "" {
		name = "cfg.json"
	}
	path := filepath.Join(dir, name)
	b, err := json.Marshal(data)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, b, 0o600))
	return path
}

func Test_parseJson_SourcesAndPrecedence(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	dir := t.TempDir()
	pathFlag := writeTempJSON(t, dir, "flag.json", map[string]any{
		"endpoint_addr_grpc":              "www.example:9000",
		"database_dsn":                    "vault.db",
		"secret_key":                      "my_secret_key",
		"access_token_validity_duration":  "1m",
		"refresh_token_validity_duration": "3m",
		"s3_root_user":                    "user",
		"s3_root_password":                "password",
		"s3_bucket":                       "bucket",
		"s3_region":                       "region",
		"s3_base_endpoint":                "base_endpoint",
	})

	t.Run("loads from json", func(t *testing.T) {
		os.Args = []string{"testbin", "-config", pathFlag}

		cfg := &Config{}
		parseJson(cfg)

		assert.Equal(t, "www.example:9000", cfg.EndpointAddrGRPC)
		assert.Equal(t, "vault.db", cfg.DatabaseDSN)
		assert.Equal(t, "my_secret_key", cfg.SecretKey)
		assert.Equal(t, 1*time.Minute, cfg.AccessTokenValidityDuration)
		assert.Equal(t, 3*time.Minute, cfg.RefreshTokenValidityDuration)
		assert.Equal(t, "user", cfg.S3RootUser)
		assert.Equal(t, "password", cfg.S3RootPassword)
		assert.Equal(t, "bucket", cfg.S3Bucket)
		assert.Equal(t, "region", cfg.S3Region)
		assert.Equal(t, "base_endpoint", cfg.S3BaseEndpoint)
	})

	t.Run("no CONFIG and no flags → no changes", func(t *testing.T) {
		os.Args = []string{"testbin"}

		cfg := &Config{
			EndpointAddrGRPC:             "defaults:1234",
			DatabaseDSN:                  "vault.db",
			SecretKey:                    "key",
			AccessTokenValidityDuration:  2 * time.Minute,
			RefreshTokenValidityDuration: 3 * time.Minute,
			S3RootUser:                   "s3root",
			S3RootPassword:               "s3rootpassword",
			S3Bucket:                     "s3bucket",
			S3Region:                     "s3region",
			S3BaseEndpoint:               "s3baseendpoint",
		}
		parseJson(cfg)

		assert.Equal(t, "defaults:1234", cfg.EndpointAddrGRPC)
		assert.Equal(t, "vault.db", cfg.DatabaseDSN)
		assert.Equal(t, "key", cfg.SecretKey)
		assert.Equal(t, 2*time.Minute, cfg.AccessTokenValidityDuration)
		assert.Equal(t, 3*time.Minute, cfg.RefreshTokenValidityDuration)
		assert.Equal(t, "s3root", cfg.S3RootUser)
		assert.Equal(t, "s3rootpassword", cfg.S3RootPassword)
		assert.Equal(t, "s3bucket", cfg.S3Bucket)
		assert.Equal(t, "s3region", cfg.S3Region)
		assert.Equal(t, "s3baseendpoint", cfg.S3BaseEndpoint)
	})

	t.Run("invalid JSON → panics", func(t *testing.T) {
		bad := filepath.Join(dir, "bad.json")
		require.NoError(t, os.WriteFile(bad, []byte(`{ this is not valid json`), 0o600))

		os.Args = []string{"testbin", "-config", bad}

		cfg := &Config{}
		require.Panics(t, func() { parseJson(cfg) })
	})
}
