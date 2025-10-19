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

// args := flagx.FilterArgs(os.Args[1:], []string{"-a", "-d", "-s", "-t", "-r", "-u", "-p", "-b", "-g", "-e"})

func TestParseFlags(t *testing.T) {

	// Test cases
	tests := []struct {
		expected    *Config
		name        string
		args        []string
		expectPanic bool
	}{
		{name: "Test1 OK", args: []string{"cmd",
			"-a", "127.0.0.1:9090", "-d", "db", "-s", "secret",
			"-t", "1", "-r", "3", "-u", "user", "-p", "password", "-b", "bucket", "-g", "us-west-1", "-e", "http://endpoint",
		}, expectPanic: false,
			expected: &Config{
				EndpointAddrGRPC:             "127.0.0.1:9090",
				DatabaseDSN:                  "db",
				SecretKey:                    "secret",
				AccessTokenValidityDuration:  1 * time.Minute,
				RefreshTokenValidityDuration: 3 * time.Minute,
				S3RootUser:                   "user",
				S3RootPassword:               "password",
				S3Bucket:                     "bucket",
				S3Region:                     "us-west-1",
				S3BaseEndpoint:               "http://endpoint",
			}},
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
