package cli

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSimpleText(t *testing.T) {
	in := bufio.NewReader(strings.NewReader("hello world\n"))
	var out bytes.Buffer
	got, err := GetSimpleText(in, "Name?", &out)
	if err != nil || got != "hello world" {
		t.Fatalf("got %q, err=%v", got, err)
	}
}

func TestGetSimpleTextEOF(t *testing.T) {
	in := bufio.NewReader(strings.NewReader("lastline"))
	var out bytes.Buffer
	got, err := GetSimpleText(in, "Name?", &out)
	if err != nil || got != "lastline" {
		t.Fatalf("got %q, err=%v", got, err)
	}
}

func TestGetMultiline_DoubleEnter(t *testing.T) {
	in := bufio.NewReader(strings.NewReader("a\nb\n\n\n"))
	var out bytes.Buffer
	got, err := GetMultiline(in, "Enter text", &out)
	if err != nil {
		t.Fatal(err)
	}
	want := "a\nb"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGetPassword_Error(t *testing.T) {
	old := readPassword
	defer func() { readPassword = old }()
	readPassword = func(int) ([]byte, error) {
		return nil, errors.New("boom")
	}
	var out bytes.Buffer
	_, err := GetPassword(&out)
	if err == nil {
		t.Fatal("expected error")
	}
}

func rdr(s string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(s))
}

func TestGetMetadata(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Unix newlines, stop on empty line",
			input:    "a=1\nb=2\n\n",
			expected: []string{"a=1", "b=2"},
		},
		{
			name:     "Windows CRLF, stop on empty line",
			input:    "a=1\r\nb=2\r\n\r\n",
			expected: []string{"a=1", "b=2"},
		},
		{
			name:     "Immediate blank line gives empty slice",
			input:    "\n",
			expected: []string{},
		},
		{
			name:     "EOF without trailing blank line",
			input:    "a=1\nb=2",
			expected: []string{"a=1", "b=2"},
		},
		{
			name:     "Spaces are preserved (no trimming except CR/LF)",
			input:    " name = value \n\n",
			expected: []string{" name = value "},
		},
		{
			name:     "Line with only spaces is considered non-empty and kept",
			input:    "   \n\n",
			expected: []string{"   "},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := GetMetadata(rdr(tc.input))
			require.NoError(t, err)
			require.Equal(t, tc.expected, got)
		})
	}
}
