package cli

import (
	"bufio"
	"context"
	"fmt"
	"strings"
)

var printlnFn = fmt.Println

type execIface interface {
	isLoggedIn() bool
	Register(ctx context.Context) error
	Login(ctx context.Context) error
	AddNote(ctx context.Context) error
	List(ctx context.Context) error
	AddLogin(ctx context.Context) error
	AddFile(ctx context.Context) error
	AddCreditCard(ctx context.Context) error
	Show(ctx context.Context) error
	Sync(ctx context.Context) error
	Logout(ctx context.Context) error
}

func runREPL(ctx context.Context, a execIface, statusFn func() string, scanner *bufio.Scanner) {
	for {
		printlnFn(fmt.Sprintf("gk> %s > ", statusFn()))
		if !scanner.Scan() {
			return
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
				printlnFn("Available commands: (l)ist, addnote, addlogin, logout, exit, show")
			} else {
				printlnFn("Available commands: register, login")
			}
		case "register":
			_ = a.Register(ctx)
		case "login":
			_ = a.Login(ctx)
		case "addnote":
			_ = a.AddNote(ctx)
		case "addlogin":
			_ = a.AddLogin(ctx)
		case "addfile":
			_ = a.AddFile(ctx)
		case "addcard":
			_ = a.AddCreditCard(ctx)
		case "show":
			_ = a.Show(ctx)
		case "l", "list":
			_ = a.List(ctx)
		case "sync":
			_ = a.Sync(ctx)
		case "exit", "quit":
			printlnFn("Bye!")
			return
		case "get":
			if len(args) == 0 {
				printlnFn("Usage: get <id>")
				continue
			}
			printlnFn(fmt.Sprintf("fetching entry %s ... (stub)", args[0]))
		case "logout":
			_ = a.Logout(ctx)
		default:
			printlnFn("Unknown command:", cmd)
		}
	}
}
