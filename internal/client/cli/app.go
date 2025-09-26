package cli

import (
	"bufio"
	"context"
	"log"
	"os"

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

	repositories, err := client.InitDatabase(ctx, "vault.db")
	if err != nil {
		log.Printf("error initializing database: %s", err.Error())
		return nil, err
	}

	apiClient, err := client.NewGophKeeperClientService(c.ServerEndpointAddr)
	if err != nil {
		return nil, err
	}

	as := services.NewAuthService(apiClient, repositories.Metadata)
	if err != nil {
		return nil, err
	}

	es := services.NewEntryService(apiClient, repositories.Entry)
	if err != nil {
		return nil, err
	}

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
	return a.userName != ""
}
