package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlags(t *testing.T) {

	// Test cases
	tests := []struct {
		expected    *Config
		name        string
		args        []string
		expectPanic bool
	}{
		{name: "Test1 OK", args: []string{"cmd", "-a", "127.0.0.1:9090", "-i", "10"}, expectPanic: false,
			expected: &Config{ServerEndpointAddr: "127.0.0.1:9090", OnlineCheckInterval: 10 * time.Second}},
		{name: "Test2 incorrect check interval", args: []string{"cmd", "-a", "127.0.0.1:9090", "-i", "abc"}, expectPanic: true, expected: &Config{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.PanicOnError)

			os.Args = tt.args

			config := &Config{}

			if !tt.expectPanic {

				require.NotPanics(t, func() { parseFlags(config) })
				assert.Empty(t, cmp.Diff(config, tt.expected))
			} else {
				require.Panics(t, func() { parseFlags(config) })
			}
		})
	}
}
