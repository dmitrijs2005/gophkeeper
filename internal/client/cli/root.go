package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
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
	printlnFn("Welcome to GophKeeper CLI (type 'help' for commands)")
	scanner := bufio.NewScanner(os.Stdin)

	a.Login(ctx) // как у тебя было

	go func() {
		a.StartOnlineStatusWatcher(ctx, a.config.OnlineCheckInterval)
	}()

	runREPL(ctx, a, a.getStatus, scanner)
}
