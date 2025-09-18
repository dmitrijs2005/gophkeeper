package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func (a *App) Register() {

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

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = a.clientService.Register(context.Background(), userName, password)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Success!")

}
