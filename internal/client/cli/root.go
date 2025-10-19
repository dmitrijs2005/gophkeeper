package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
)

// getStatus builds a short status suffix for the REPL prompt.
//
// Format examples:
//
//	""                      -> ""         (no user, no mode)
//	user="", mode="online"  -> "(online)"
//	user="alice", mode=""   -> "(alice)"
//	user="alice", mode="online" -> "(alice online)"
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

// Root is the entrypoint for the interactive CLI session.
//
// It prints a welcome banner, logs the user in (interactive), starts a
// background online-status watcher, and then hands control to the REPL.
// The call blocks until the REPL returns (e.g., user types "exit" or EOF).
func (a *App) Root(ctx context.Context) {
	printlnFn("Welcome to GophKeeper CLI (type 'help' for commands)")
	scanner := bufio.NewScanner(os.Stdin)

	a.Login(ctx)

	go func() {
		a.StartOnlineStatusWatcher(ctx, a.config.OnlineCheckInterval)
	}()

	runREPL(ctx, a, a.getStatus, scanner)
}
