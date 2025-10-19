package cli

import (
	"bufio"
	"context"
	"errors"
	"io"
	"testing"
)

func stubPassword1(t *testing.T, pw []byte) func() {
	t.Helper()
	orig := getPassword
	getPassword = func(_ io.Writer) ([]byte, error) { return pw, nil }
	return func() { getPassword = orig }
}

func stubInputs(t *testing.T, username string, password []byte) func() {
	t.Helper()
	origST, origGP := getSimpleText, getPassword
	getSimpleText = func(_ *bufio.Reader, _ string, _ io.Writer) (string, error) { return username, nil }
	getPassword = func(_ io.Writer) ([]byte, error) { return password, nil }
	return func() {
		getSimpleText = origST
		getPassword = origGP
	}
}

type fakeAuth struct {
	// Register
	regUser string
	regPass []byte
	regErr  error

	// OnlineLogin
	onlineUser string
	onlinePass []byte
	onlineMK   []byte
	onlineErr  error

	// OfflineLogin
	offlineUser string
	offlinePass []byte
	offlineMK   []byte
	offlineErr  error

	// ClearOfflineData
	clearCalled bool
	clearErr    error
}

func (f *fakeAuth) Register(_ context.Context, user string, pass []byte) error {
	f.regUser, f.regPass = user, append([]byte(nil), pass...)
	return f.regErr
}
func (f *fakeAuth) OnlineLogin(_ context.Context, user string, pass []byte) ([]byte, error) {
	f.onlineUser, f.onlinePass = user, append([]byte(nil), pass...)
	return f.onlineMK, f.onlineErr
}
func (f *fakeAuth) OfflineLogin(_ context.Context, user string, pass []byte) ([]byte, error) {
	f.offlineUser, f.offlinePass = user, append([]byte(nil), pass...)
	return f.offlineMK, f.offlineErr
}
func (f *fakeAuth) ClearOfflineData(context.Context) error {
	f.clearCalled = true
	return f.clearErr
}
func (f *fakeAuth) Close(ctx context.Context) error { return nil }
func (f *fakeAuth) Ping(ctx context.Context) error  { return nil }

func TestRegister_Success(t *testing.T) {
	f := &fakeAuth{}
	a := &App{authService: f}

	restore := stubInputs(t, "alice@example.org", []byte("secret"))
	defer restore()

	if err := a.Register(context.Background()); err != nil {
		t.Fatalf("Register err: %v", err)
	}
	if f.regUser != "alice@example.org" {
		t.Fatalf("Register user mismatch: %q", f.regUser)
	}
	if string(f.regPass) != "secret" {
		t.Fatalf("Register pass mismatch: %q", string(f.regPass))
	}
}

func TestLogout(t *testing.T) {
	f := &fakeAuth{}
	a := &App{authService: f, masterKey: []byte("something")}
	if err := a.Logout(context.Background()); err != nil {
		t.Fatalf("Logout err: %v", err)
	}
	if !f.clearCalled {
		t.Fatalf("ClearOfflineData not called")
	}
	if a.masterKey != nil {
		t.Fatalf("masterKey not cleared")
	}
}

func TestLogout_ErrorPropagates(t *testing.T) {
	f := &fakeAuth{clearErr: errors.New("clean-fail")}
	a := &App{authService: f}
	if err := a.Logout(context.Background()); err == nil {
		t.Fatalf("want error from ClearOfflineData")
	}
}
