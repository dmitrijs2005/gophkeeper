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

type Mode string

const (
	ModeOffline  Mode = "offline"
	ModeOnline   Mode = "online"
	ModeDisabled Mode = "disabled"
)

type App struct {
	config       *config.Config
	authService  services.AuthService
	entryService services.EntryService
	masterKey    []byte
	userName     string
	Mode         Mode
	reader       *bufio.Reader
}

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

	return &App{config: c, authService: as, entryService: es, reader: bufio.NewReader(os.Stdin)}, nil
}

func (app *App) setMode(mode Mode) {
	if app.Mode != mode {
		app.Mode = mode
		log.Printf("Switched to %s mode\n", mode)
	}
}

func (a *App) Run(ctx context.Context) {
	defer a.authService.Close(ctx)
	a.Root(ctx)
}

func (a *App) isLoggedIn() bool {
	return a.masterKey != nil
}

func (a *App) StartOnlineStatusWatcher(ctx context.Context, interval time.Duration) {

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			err := a.authService.Ping(ctx)
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
