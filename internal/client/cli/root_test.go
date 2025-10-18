package cli

import (
	"bufio"
	"context"
	"io"
	"strings"
	"testing"
)

func stubPassword(t *testing.T, pw []byte) func() {
	t.Helper()
	orig := getPassword
	getPassword = func(_ io.Writer) ([]byte, error) { return pw, nil }
	return func() { getPassword = orig }
}

// ---- getStatus ----

func TestGetStatus_Empty(t *testing.T) {
	a := &App{}
	got := a.getStatus()
	if got != "" {
		t.Fatalf("want empty status, got %q", got)
	}
}

func TestGetStatus_WithUsernameOnly(t *testing.T) {
	a := &App{userName: "alice"}
	got := a.getStatus()
	want := "(alice )"
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

// ---- runREPL (smoke) ----

func silencePrintln(t *testing.T) {
	t.Helper()
	orig := printlnFn
	printlnFn = func(...any) (int, error) { return 0, nil }
	t.Cleanup(func() { printlnFn = orig })
}

type fakeExec1 struct {
	logged bool
}

func (f *fakeExec1) isLoggedIn() bool                    { return f.logged }
func (f *fakeExec1) Register(context.Context) error      { return nil }
func (f *fakeExec1) Login(context.Context) error         { f.logged = true; return nil }
func (f *fakeExec1) AddNote(context.Context) error       { return nil }
func (f *fakeExec1) List(context.Context) error          { return nil }
func (f *fakeExec1) AddLogin(context.Context) error      { return nil }
func (f *fakeExec1) AddFile(context.Context) error       { return nil }
func (f *fakeExec1) AddCreditCard(context.Context) error { return nil }
func (f *fakeExec1) Show(context.Context) error          { return nil }
func (f *fakeExec1) Sync(context.Context) error          { return nil }
func (f *fakeExec1) Logout(context.Context) error        { f.logged = false; return nil }

func TestRunREPL_HelpThenQuit(t *testing.T) {
	silencePrintln(t)

	input := "help\nquit\n"
	sc := bufio.NewScanner(strings.NewReader(input))

	exec := &fakeExec1{}
	status := func() string { return "status" }

	runREPL(context.Background(), exec, status, sc)
}
