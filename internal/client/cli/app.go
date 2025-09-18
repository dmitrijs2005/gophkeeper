package cli

import (
	"fmt"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/client/config"
	"github.com/dmitrijs2005/gophkeeper/internal/client/service"
)

type App struct {
	config        *config.Config
	clientService service.Service
	masterKey     []byte
	userName      string
	lastActivity  time.Time
}

func NewApp(c *config.Config) (*App, error) {

	s, err := service.NewGophKeeperClientService(c.ServerEndpointAddr)
	if err != nil {
		return nil, err
	}

	err = s.InitGRPCClient()
	if err != nil {
		return nil, err
	}

	return &App{config: c, clientService: s}, nil
}

func (a *App) Run() {

	defer a.clientService.Close()
	a.Main()
}

func (a *App) isLoggedIn() bool {
	return a.userName != ""
}

func (a *App) showLogin() string {
	if !a.isLoggedIn() {
		return ""
	}
	return fmt.Sprintf("(%s)", a.userName)
}
