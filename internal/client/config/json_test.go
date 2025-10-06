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
		"server_endpoint_addr":  "www.example:9000",
		"online_check_interval": "10s",
	})

	t.Run("loads from flags", func(t *testing.T) {
		os.Args = []string{"testbin", "-config", pathFlag}

		cfg := &Config{}
		parseJson(cfg)

		assert.Equal(t, "www.example:9000", cfg.ServerEndpointAddr)
		assert.Equal(t, 10*time.Second, cfg.OnlineCheckInterval)
	})

	t.Run("no CONFIG and no flags → no changes", func(t *testing.T) {
		os.Args = []string{"testbin"}

		cfg := &Config{
			ServerEndpointAddr:  "defaults:1234",
			OnlineCheckInterval: 42 * time.Second,
		}
		parseJson(cfg)

		assert.Equal(t, "defaults:1234", cfg.ServerEndpointAddr)
		assert.Equal(t, 42*time.Second, cfg.OnlineCheckInterval)
	})

	t.Run("invalid JSON → panics", func(t *testing.T) {
		bad := filepath.Join(dir, "bad.json")
		require.NoError(t, os.WriteFile(bad, []byte(`{ this is not valid json`), 0o600))

		os.Args = []string{"testbin", "-config", bad}

		cfg := &Config{}
		require.Panics(t, func() { parseJson(cfg) })
	})
}
