package cli

import (
	"bufio"
	"context"
	"fmt"
	"strings"
)

// printlnFn is a test seam for user-facing output. In tests, replace it with a stub.
var printlnFn = fmt.Println

// execIface defines the minimal command surface the REPL needs to operate.
// The real App type satisfies this interface; tests can provide a lightweight stub.
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

// runREPL starts a simple read–eval–print loop for the GophKeeper CLI.
//
// It reads a line from the provided scanner, parses the first token as the
// command, and dispatches to methods on 'a'. Unknown commands are reported
// back to the user. The loop exits on scanner EOF or when the user types
// "exit" or "quit".
//
// Prompt & Commands
//
// The prompt shows the current status (from statusFn) and accepts commands:
//
//	Not logged in:
//	  - help           — show available commands
//	  - register       — create an account
//	  - login          — authenticate
//	  - exit | quit    — leave the program
//
//	Logged in:
//	  - help           — show available commands
//	  - addnote        — add a note
//	  - addlogin       — add login credentials
//	  - addfile        — add a binary file
//	  - addcard        — add a credit card
//	  - list       	   — list entries
//	  - show           — show a single entry (interactive ID prompt)
//	  - sync           — synchronize with the server
//	  - logout         — log out
//	  - exit | quit    — leave the program
//
// Any errors returned by command handlers are ignored here; handlers should
// log their own errors. This keeps the REPL loop resilient and focused on I/O.
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

		switch cmd {
		case "help":
			if a.isLoggedIn() {
				printlnFn("Available commands: (l)ist, addnote, addlogin, addfile, addcard, show, sync, logout, exit")
			} else {
				printlnFn("Available commands: register, login, exit")
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

		case "logout":
			_ = a.Logout(ctx)

		case "exit", "quit":
			printlnFn("Bye!")
			return

		default:
			printlnFn("Unknown command:", cmd)
		}
	}
}
