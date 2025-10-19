package cli

import (
	"bytes"
	"log"
	"testing"
)

func TestIsLoggedIn_NilMasterKey(t *testing.T) {
	app := &App{}
	if app.isLoggedIn() {
		t.Fatalf("expected isLoggedIn() == false when masterKey is nil")
	}
}

func TestIsLoggedIn_NonNilMasterKey(t *testing.T) {
	app := &App{masterKey: []byte{1, 2, 3}}
	if !app.isLoggedIn() {
		t.Fatalf("expected isLoggedIn() == true when masterKey is not nil")
	}
}

func TestSetMode_ChangesAndLogsOnce(t *testing.T) {
	app := &App{}
	var buf bytes.Buffer

	old := log.Default().Writer()
	defer log.SetOutput(old)
	log.SetOutput(&buf)

	app.setMode(ModeOnline)
	if app.Mode != ModeOnline {
		t.Fatalf("expected mode to be %q, got %q", ModeOnline, app.Mode)
	}
	if got := buf.String(); got == "" {
		t.Fatalf("expected log output on mode change, got empty")
	}

	buf.Reset()

	app.setMode(ModeOnline)
	if app.Mode != ModeOnline {
		t.Fatalf("expected mode to remain %q, got %q", ModeOnline, app.Mode)
	}
	if got := buf.String(); got != "" {
		t.Fatalf("expected no log output when mode doesn't change, got: %q", got)
	}

	app.setMode(ModeOffline)
	if app.Mode != ModeOffline {
		t.Fatalf("expected mode to be %q, got %q", ModeOffline, app.Mode)
	}
	if got := buf.String(); got == "" {
		t.Fatalf("expected log output on mode change to offline, got empty")
	}
}
