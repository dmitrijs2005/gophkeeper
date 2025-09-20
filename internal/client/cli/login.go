package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dmitrijs2005/gophkeeper/internal/client/utils"
	"github.com/dmitrijs2005/gophkeeper/internal/shared"
	"golang.org/x/term"
)

func (a *App) Login() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter user name (email)")

	userName, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	userName = strings.TrimSpace(userName)

	fmt.Println("Enter password")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	defer shared.WipeByteArray(password)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	salt, err := a.clientService.GetSalt(context.Background(), userName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	key := utils.DeriveMasterKey(password, salt)

	err = a.clientService.Login(context.Background(), userName, key)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Success!")

	a.userName = userName
	a.masterKey = key

}
