package cli

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"testing"
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
