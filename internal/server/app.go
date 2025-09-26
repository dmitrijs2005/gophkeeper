// Package server initializes and runs the main application server.
// It configures storage backends, handles graceful shutdown, restores and saves metric dumps,
// and starts the HTTP server for metric collection.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dmitrijs2005/gophkeeper/internal/logging"
	"github.com/dmitrijs2005/gophkeeper/internal/server/config"
	"github.com/dmitrijs2005/gophkeeper/internal/server/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/server/shared/db"
	"github.com/dmitrijs2005/gophkeeper/internal/server/users"

	gs "github.com/dmitrijs2005/gophkeeper/internal/server/grpc"
)

type App struct {
	config       *config.Config
	logger       logging.Logger
	userService  *users.Service
	entryService *entries.Service
}

func NewApp(c *config.Config) (*App, error) {

	slog := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger := logging.NewSlogLogger(slog)

	config := config.LoadConfig()

	um, err := db.NewPostgresRepositoryManager(config.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("db init error: %w", err)
	}

	us := users.NewService(um.Users(), um.RefreshTokens(), c)
	es := entries.NewService(um.Entries(), c)

	return &App{config: config, logger: logger, userService: us, entryService: es}, nil
}

func (app *App) initSignalHandler(cancelFunc context.CancelFunc) {
	// Channel to catch OS signals.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigs
		cancelFunc()
	}()
}

func (app *App) startGRPCServer(ctx context.Context, cancelFunc context.CancelFunc) {

	s, err := gs.NewgGRPCServer(app.config.EndpointAddrGRPC, app.logger, app.userService, app.entryService, app.config.SecretKey)

	if err != nil {
		app.logger.Error(ctx, err.Error())
		cancelFunc()
	} else {

		if err := s.Run(ctx); err != nil {
			app.logger.Error(ctx, err.Error())
			cancelFunc()
		}
	}
}

func (app *App) Run(ctx context.Context) {

	ctx, cancelFunc := context.WithCancel(context.Background())

	app.logger.Info(ctx, "Starting app...")

	app.initSignalHandler(cancelFunc)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.startGRPCServer(ctx, cancelFunc)
	}()

	wg.Wait()

}

// func (app *App) closeDBIfNeeded(s storage.Storage) (bool, error) {

// 	db, ok := s.(storage.DBStorage)
// 	if ok {
// 		err := db.Close()
// 		return true, err
// 	}

// 	return false, nil

// }
