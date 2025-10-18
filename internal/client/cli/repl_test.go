package cli

import (
	"bufio"
	"context"
	"strings"
	"testing"
)

type fakeExec struct {
	loggedIn bool

	calls []string
	arg   string
}

func (f *fakeExec) isLoggedIn() bool { return f.loggedIn }
func (f *fakeExec) Register(ctx context.Context) error {
	f.calls = append(f.calls, "register")
	return nil
}
func (f *fakeExec) Login(ctx context.Context) error {
	f.calls = append(f.calls, "login")
	f.loggedIn = true
	return nil
}
func (f *fakeExec) AddNote(ctx context.Context) error {
	f.calls = append(f.calls, "addnote")
	return nil
}
func (f *fakeExec) List(ctx context.Context) error { f.calls = append(f.calls, "list"); return nil }
func (f *fakeExec) AddLogin(ctx context.Context) error {
	f.calls = append(f.calls, "addlogin")
	return nil
}
func (f *fakeExec) AddFile(ctx context.Context) error {
	f.calls = append(f.calls, "addfile")
	return nil
}
func (f *fakeExec) AddCreditCard(ctx context.Context) error {
	f.calls = append(f.calls, "addcard")
	return nil
}
func (f *fakeExec) Show(ctx context.Context) error {
	f.calls = append(f.calls, "show")
	return nil
}
func (f *fakeExec) Sync(ctx context.Context) error { f.calls = append(f.calls, "sync"); return nil }
func (f *fakeExec) Logout(ctx context.Context) error {
	f.calls = append(f.calls, "logout")
	f.loggedIn = false
	return nil
}

func TestRunREPL_LoginFlowAndCommands(t *testing.T) {
	origPrint := printlnFn
	printlnFn = func(...any) (int, error) { return 0, nil }
	t.Cleanup(func() { printlnFn = origPrint })

	input := strings.NewReader(strings.Join([]string{
		"help",
		"login",
		"help",
		"addnote",
		"list",
		"show 123",
		"sync",
		"get 42",
		"foobar",
		"exit",
	}, "\n"))

	exec := &fakeExec{loggedIn: false}
	sc := bufio.NewScanner(input)

	runREPL(context.Background(), exec, func() string { return "status" }, sc)

	wantOrder := []string{"login", "addnote", "list", "show", "sync"}
	if len(exec.calls) < len(wantOrder) {
		t.Fatalf("few calls: %+v", exec.calls)
	}
	idx := 0
	for _, c := range exec.calls {
		if idx < len(wantOrder) && c == wantOrder[idx] {
			idx++
		}
	}
	if idx != len(wantOrder) {
		t.Fatalf("commands order mismatch: got %v, want subseq %v", exec.calls, wantOrder)
	}

}

func TestRunREPL_UsageAndQuit(t *testing.T) {
	origPrint := printlnFn
	printlnFn = func(...any) (int, error) { return 0, nil }
	t.Cleanup(func() { printlnFn = origPrint })

	input := strings.NewReader("get\nquit\n")
	exec := &fakeExec{loggedIn: true}
	sc := bufio.NewScanner(input)

	runREPL(context.Background(), exec, func() string { return "s" }, sc)

	if len(exec.calls) != 0 {
		t.Fatalf("unexpected calls: %v", exec.calls)
	}
}
