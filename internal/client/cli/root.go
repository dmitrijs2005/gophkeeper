package cli

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
)

func (a *App) getStatus() string {
	s := ""
	if a.userName != "" {
		s = a.userName + " "
	}
	if a.Mode != "" {
		s = s + string(a.Mode)
	}
	if s != "" {
		s = fmt.Sprintf("(%s)", s)
	}
	return s
}

func (a *App) Root(ctx context.Context) {

	log.Println("Welcome to GophKeeper CLI (type 'help' for commands)")
	scanner := bufio.NewScanner(os.Stdin)

	a.Login(ctx)

	go func() {
		a.StartOnlineStatusWatcher(ctx, a.config.OnlineCheckInterval)
	}()

	for {
		fmt.Printf("gcli %s> ", a.getStatus())
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]
		args := parts[1:]

		switch cmd {
		case "help":
			if a.isLoggedIn() {
				fmt.Println("Available commands: (l)ist, addnote, addlogin, logout, exit, show")
			} else {
				fmt.Println("Available commands: register, login")
			}

		case "register":
			a.Register(ctx)
		case "login":
			a.Login(ctx)
		case "addnote":
			a.addNote(ctx)
		case "addlogin":
			a.addLogin(ctx)
		case "list":
			a.list(ctx)
		case "delete":
			a.delete(ctx)
		case "addfile":
			a.addFile(ctx)
		case "addcard":
			a.addCreditCard(ctx)
		case "show":
			a.show(ctx)
		case "sync":
			a.sync(ctx)
		case "exit", "quit":
			fmt.Println("Bye!")
			return
		case "get":
			if len(args) == 0 {
				fmt.Println("Usage: get <id>")
				continue
			}
			fmt.Printf("Fetching entry %s ... (stub)\n", args[0])
		case "logout":
			a.Logout(ctx)
		default:
			fmt.Println("Unknown command:", cmd)
		}
	}

}
