package buildinfo

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintBuildParam(t *testing.T) {
	tests := []struct {
		name     string
		param    string
		value    string
		expected string
	}{
		{"WithValue", "Build version", "1.2.3", "Build version: 1.2.3\n"},
		{"WithoutValue", "Build version", "", "Build version: N/A\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			printBuildParam(&buf, tt.param, tt.value)

			if got := buf.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func Test_printBuildParam(t *testing.T) {
	t.Run("non-empty value", func(t *testing.T) {
		var buf bytes.Buffer
		printBuildParam(&buf, "Build version", "1.2.3")
		assert.Equal(t, "Build version: 1.2.3\n", buf.String())
	})

	t.Run("empty value -> N/A", func(t *testing.T) {
		var buf bytes.Buffer
		printBuildParam(&buf, "Build date", "")
		assert.Equal(t, "Build date: N/A\n", buf.String())
	})
}

func Test_PrintBuildData_AllEmpty(t *testing.T) {
	// Save & restore globals
	oldV, oldD, oldC := buildVersion, buildDate, buildCommit
	t.Cleanup(func() { buildVersion, buildDate, buildCommit = oldV, oldD, oldC })

	buildVersion, buildDate, buildCommit = "", "", ""

	var buf bytes.Buffer
	PrintBuildData(&buf)

	require.Equal(t,
		"Build version: N/A\nBuild date: N/A\nBuild commit: N/A\n",
		buf.String(),
	)
}

func Test_PrintBuildData_CustomValues(t *testing.T) {
	// Save & restore globals
	oldV, oldD, oldC := buildVersion, buildDate, buildCommit
	t.Cleanup(func() { buildVersion, buildDate, buildCommit = oldV, oldD, oldC })

	buildVersion = "v1.0.0"
	buildDate = "2025-08-18T12:34:56Z"
	buildCommit = "abcdef123456"

	var buf bytes.Buffer
	PrintBuildData(&buf)

	want := "Build version: v1.0.0\n" +
		"Build date: 2025-08-18T12:34:56Z\n" +
		"Build commit: abcdef123456\n"

	require.Equal(t, want, buf.String())
}
