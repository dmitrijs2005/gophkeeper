package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	var c Config
	c.LoadDefaults()

	assert.Equal(t, "127.0.0.1:50051", c.ServerEndpointAddr)
	assert.Equal(t, 3*time.Second, c.OnlineCheckInterval)
}

func TestLoadConfig_UsesDefaultsBeforeParsing(t *testing.T) {
	cfg := LoadConfig()

	require.NotNil(t, cfg, "LoadConfig must not return nil")
	assert.Equal(t, "127.0.0.1:50051", cfg.ServerEndpointAddr)
	assert.Equal(t, 3*time.Second, cfg.OnlineCheckInterval)
}
