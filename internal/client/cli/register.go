package cli

import (
	"context"
	"fmt"
	"log"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
)

func (a *App) Register(ctx context.Context) {

	userName, err := GetSimpleText(a.reader, "-Enter email")
	if err != nil {
		log.Printf("error: %v", err)
	}

	password, err := GetPassword()
	if err != nil {
		log.Printf("error: %v", err)
	}

	defer common.WipeByteArray(password)

	err = a.authService.Register(ctx, userName, password)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Success!")

}
