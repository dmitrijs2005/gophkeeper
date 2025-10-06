package flagx

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		allowedFlags []string
		want         []string
	}{
		{
			name:         "short flag with separate value",
			args:         []string{"-c", "conf.json", "-a", "localhost"},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{"-c", "conf.json"},
		},
		{
			name:         "long flag with equals",
			args:         []string{"--config=alt.json", "-a", "localhost"},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{"--config=alt.json"},
		},
		{
			name:         "both short and long present, preserve order",
			args:         []string{"--config=first.json", "-c", "second.json", "-x", "1"},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{"--config=first.json", "-c", "second.json"},
		},
		{
			name:         "unknown flags ignored",
			args:         []string{"-x", "1", "--y=2", "positional"},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{},
		},
		{
			name:         "flag without value at end is kept as-is",
			args:         []string{"-c"},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{"-c"},
		},
		{
			name:         "flag followed by another flag (no value)",
			args:         []string{"-c", "-notvalue"},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{"-c"},
		},
		{
			name:         "value that looks like a flag but with equals form",
			args:         []string{"--config=--weird.json"},
			allowedFlags: []string{"--config"},
			want:         []string{"--config=--weird.json"},
		},
		{
			name:         "multiple allowed flags kept",
			args:         []string{"-a", "localhost:8080", "-c", "conf.json", "--other", "x"},
			allowedFlags: []string{"-c", "-a"},
			want:         []string{"-a", "localhost:8080", "-c", "conf.json"},
		},
		{
			name:         "empty args",
			args:         []string{},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{},
		},
		{
			name:         "path with spaces remains single arg",
			args:         []string{"-c", "/home/user/conf.json"},
			allowedFlags: []string{"-c"},
			want:         []string{"-c", "/home/user/conf.json"},
		},
		{
			name:         "do not treat next dash-starting token as value",
			args:         []string{"-c", "--config=alt.json"},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{"-c", "--config=alt.json"},
		},
		{
			name:         "repeated allowed flag is preserved in order",
			args:         []string{"-c", "one.json", "-c", "two.json"},
			allowedFlags: []string{"-c"},
			want:         []string{"-c", "one.json", "-c", "two.json"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := FilterArgs(tt.args, tt.allowedFlags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("FilterArgs() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_jsonConfigFlags(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	t.Run("short -c with value", func(t *testing.T) {
		os.Args = []string{"testbin", "-c", "/path/short.json"}
		assert.Equal(t, "/path/short.json", JsonConfigFlags())
	})

	t.Run("long -config with value", func(t *testing.T) {
		os.Args = []string{"testbin", "-config", "/path/long.json"}
		assert.Equal(t, "/path/long.json", JsonConfigFlags())
	})

	t.Run("unknown flags are ignored", func(t *testing.T) {
		os.Args = []string{"testbin", "-x", "1", "-y", "2"}
		assert.Empty(t, JsonConfigFlags())
	})

	t.Run("multiple flags, last wins", func(t *testing.T) {
		os.Args = []string{"testbin", "-c", "/path/1.json", "-config", "/path/2.json"}
		assert.Equal(t, "/path/2.json", JsonConfigFlags())
	})
}
