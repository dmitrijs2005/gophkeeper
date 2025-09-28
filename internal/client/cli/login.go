package cli

import (
	"context"
	"errors"
	"log"

	"github.com/dmitrijs2005/gophkeeper/internal/client/client"
	"github.com/dmitrijs2005/gophkeeper/internal/common"
)

func (a *App) Login(ctx context.Context) {

	userName, err := GetSimpleText(a.reader, "-Enter email")
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	password, err := GetPassword()
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	defer common.WipeByteArray(password)

	var masterKey []byte
	var mode Mode

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
	}

	a.masterKey = masterKey
	a.setMode(mode)

}
