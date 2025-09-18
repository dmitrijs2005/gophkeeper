package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func (a *App) Main() {

	fmt.Println("GophKeeper CLI (type 'help' for commands)")
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("gophkeeper %s > ", a.showLogin())
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
				fmt.Println("Available commands: list, addnote, logout, exit")
			} else {
				fmt.Println("Available commands: register, login")
			}

		case "register":
			// вызвать API login, сохранить session_id
			a.Register()
		case "login":
			// вызвать API login, сохранить session_id
			a.Login()
		case "addnote":
			// вызвать API login, сохранить session_id
			a.AddNote()
		case "exit", "quit":
			fmt.Println("Bye!")
			return
		case "list":
			// запросить список записей с сервера
			fmt.Println("Entries: (stub)")
		case "get":
			if len(args) == 0 {
				fmt.Println("Usage: get <id>")
				continue
			}
			fmt.Printf("Fetching entry %s ... (stub)\n", args[0])
		case "put":
			if len(args) == 0 {
				fmt.Println("Usage: put <title>")
				continue
			}
			fmt.Printf("Creating entry with title '%s' ... (stub)\n", args[0])
		case "logout":
			fmt.Println("Logged out (stub)")

		default:
			fmt.Println("Unknown command:", cmd)
		}
	}

}
