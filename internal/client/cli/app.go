// Package cli contains the interactive command-line client for GophKeeper.
//
// The App type wires configuration, local persistence, network client services,
// and the interactive loop. It supports an online/offline mode that is kept in
// sync by a lightweight periodic connectivity check.
package cli

import (
	"bufio"
	"context"
	"log"
	"os"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/client/client"
	"github.com/dmitrijs2005/gophkeeper/internal/client/config"
	"github.com/dmitrijs2005/gophkeeper/internal/client/services"

	_ "modernc.org/sqlite"
)

// Mode represents the current connectivity state of the CLI.
type Mode string

const (
	// ModeOffline indicates the server is unreachable; only local actions
	// should be attempted.
	ModeOffline Mode = "offline"
	// ModeOnline indicates the server is reachable and online operations
	// are available.
	ModeOnline Mode = "online"
	// ModeDisabled indicates connectivity checks are turned off or not applicable.
	ModeDisabled Mode = "disabled"
)

// App is the top-level CLI application.
//
// It holds configuration, service facades, the current master encryption key
// (present when logged in), the current user name, connectivity mode, and an
// input reader for interactive prompts.
type App struct {
	// config is the runtime configuration loaded at startup.
	config *config.Config

	// authService handles authentication, tokens, and server liveness probes.
	authService services.AuthService

	// entryService manages entries (notes, logins, files) both locally and
	// via the server API.
	entryService services.EntryService

	// masterKey is set upon successful login and remains nil otherwise.
	// It should be wiped on logout.
	masterKey []byte

	// userName is the authenticated user's identifier.
	userName string

	// Mode reflects current connectivity status (online/offline/disabled).
	Mode Mode

	// reader provides interactive input for the CLI loop.
	reader *bufio.Reader
}

// NewApp constructs an App from the given config.
//
// It initializes the local SQLite database, creates an API client pointing to
// the configured server endpoint, and wires the authentication and entry
// services around them.
//
// The returned App is ready to Run. On initialization failure an error is
// returned (and may already be logged).
func NewApp(c *config.Config) (*App, error) {
	ctx := context.Background()

	db, err := client.InitDatabase(ctx, "vault.db")
	if err != nil {
		log.Printf("error initializing database: %s", err.Error())
		return nil, err
	}

	apiClient, err := client.NewGophKeeperClientService(c.ServerEndpointAddr)
	if err != nil {
		return nil, err
	}

	as := services.NewAuthService(apiClient, db)
	es := services.NewEntryService(apiClient, db)

	return &App{
		config:       c,
		authService:  as,
		entryService: es,
		reader:       bufio.NewReader(os.Stdin),
	}, nil
}

// setMode updates the connectivity Mode and logs the transition.
// It is a no-op if the mode is unchanged.
func (app *App) setMode(mode Mode) {
	if app.Mode != mode {
		app.Mode = mode
		log.Printf("Switched to %s mode\n", mode)
	}
}

// Run starts the interactive CLI.
//
// It ensures the authService is closed when the function returns and delegates
// to the Root command loop. The provided context is propagated to long-running
// operations created from within Root.
func (a *App) Run(ctx context.Context) {
	defer a.authService.Close(ctx)
	a.Root(ctx)
}

// isLoggedIn reports whether a masterKey is present (i.e., the user is logged in).
func (a *App) isLoggedIn() bool {
	return a.masterKey != nil
}

// StartOnlineStatusWatcher periodically probes server reachability and updates
// the application's Mode accordingly.
//
// A ping is attempted every 'interval'. Each probe uses its own 3-second
// timeout to avoid piling up when the server is slow or unreachable. The
// watcher stops when ctx is canceled.
func (a *App) StartOnlineStatusWatcher(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctxPing, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			err := a.authService.Ping(ctxPing)
			cancel()

			if err != nil {
				if a.Mode == ModeOnline {
					a.setMode(ModeOffline)
				}
			} else {
				if a.Mode != ModeOnline {
					a.setMode(ModeOnline)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
