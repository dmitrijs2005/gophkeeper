package cli

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/dmitrijs2005/gophkeeper/internal/client/client"
	"github.com/dmitrijs2005/gophkeeper/internal/common"
)

// getSimpleText and getPassword are indirections used to facilitate testing.
// They point to interactive input helpers and can be swapped in tests.
var getSimpleText = GetSimpleText
var getPassword = GetPassword

// Register prompts the user for an email and password and attempts to create
// a new account via the AuthService.
//
// On success it prints "Success!" and returns nil. The password byte slice
// is securely wiped before returning. Any I/O or service error is returned
// unchanged.
func (a *App) Register(ctx context.Context) error {
	userName, err := getSimpleText(a.reader, "Enter email", os.Stdout)
	if err != nil {
		return err
	}

	password, err := getPassword(os.Stdout)
	if err != nil {
		return err
	}
	defer common.WipeByteArray(password)

	if err := a.authService.Register(ctx, userName, password); err != nil {
		return err
	}

	fmt.Println("Success!")
	return nil
}

// Login prompts the user for credentials and tries to authenticate.
//
// The method first attempts an online login. If the server is unavailable
// (errors.Is(err, client.ErrUnavailable)), it falls back to offline login.
// On success it sets a.masterKey and updates connectivity Mode:
//   - ModeOnline if online login succeeds,
//   - ModeOffline if offline login succeeds,
//   - ModeDisabled if both fail.
//
// The password is securely wiped before returning. Any error from the
// underlying auth calls is returned; note that a nil error does not
// necessarily imply ModeOnline—inspect App.Mode for the final state.
func (a *App) Login(ctx context.Context) error {
	userName, err := GetSimpleText(a.reader, "Enter email", os.Stdout)
	if err != nil {
		return err
	}

	password, err := GetPassword(os.Stdout)
	if err != nil {
		return err
	}
	defer common.WipeByteArray(password)

	var (
		masterKey []byte
		mode      Mode
	)

	masterKey, err = a.authService.OnlineLogin(ctx, userName, password)
	if err != nil {
		if errors.Is(err, client.ErrUnavailable) {
			log.Printf("Server unavailable, trying offline login...")
			masterKey, err = a.authService.OfflineLogin(ctx, userName, password)
			if err != nil {
				log.Printf("Offline login unsuccessfull: %s", err.Error())
				mode = ModeDisabled
			} else {
				log.Printf("Offline login successfull")
				mode = ModeOffline
			}
		} else {
			log.Printf("Login unsuccessfull: %s", err.Error())
		}
	} else {
		log.Printf("Login successfull")
		mode = ModeOnline
	}

	a.masterKey = masterKey
	a.setMode(mode)
	return nil
}

// Logout clears locally cached offline data and removes the in-memory
// masterKey. It returns any error from the AuthService cleanup.
func (a *App) Logout(ctx context.Context) error {
	if err := a.authService.ClearOfflineData(ctx); err != nil {
		return err
	}
	a.masterKey = nil
	return nil
}
